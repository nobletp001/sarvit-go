package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nobletp001/sarvit/internal/config"
	"github.com/nobletp001/sarvit/internal/models"
	"github.com/nobletp001/sarvit/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already in use")
)

type AuthService interface {
	Signup(ctx context.Context, name, email, password string) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	ParseUserIDFromToken(tokenStr string) (primitive.ObjectID, error)
}

type authService struct {
	cfg      config.Config
	userRepo repository.UserRepository
	now      func() time.Time
}

func NewAuthService(cfg config.Config, userRepo repository.UserRepository) AuthService {
	return &authService{
		cfg:      cfg,
		userRepo: userRepo,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *authService) Signup(ctx context.Context, name, email, password string) (*models.User, string, error) {
	// check if email exists
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, "", ErrEmailExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	u := &models.User{
		ID:        primitive.NewObjectID(),
		Name:      name,
		Email:     email,
		Password:  string(hashed),
		CreatedAt: s.now(),
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, "", err
	}

	tok, err := s.signJWT(u.ID, u.Email)
	if err != nil {
		return nil, "", err
	}
	return u, tok, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	tok, err := s.signJWT(u.ID, u.Email)
	if err != nil {
		return nil, "", err
	}
	return u, tok, nil
}

func (s *authService) signJWT(userID primitive.ObjectID, email string) (string, error) {
	ttl := time.Duration(s.cfg.JWTTtlMinutes) * time.Minute
	claims := jwt.MapClaims{
		"sub": userID.Hex(),
		"eml": email,
		"iat": time.Now().UTC().Unix(),
		"exp": time.Now().UTC().Add(ttl).Unix(),
		"iss": "sarvit-api",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *authService) ParseUserIDFromToken(tokenStr string) (primitive.ObjectID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return primitive.NilObjectID, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, errors.New("invalid claims")
	}
	sub, _ := claims["sub"].(string)
	return primitive.ObjectIDFromHex(sub)
}
