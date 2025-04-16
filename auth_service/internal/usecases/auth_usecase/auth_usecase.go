package auth_usecase

import (
	"auth_service/internal/config"
	e "auth_service/internal/domain/errors"
	"auth_service/internal/domain/models"
	"auth_service/internal/repositories"
	"auth_service/internal/repositories/postgres"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type AuthUsecase struct {
	cfg           *config.Config
	log           *slog.Logger
	userRepo      repositories.UserRepository
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewAuthUsecase(
	cfg *config.Config,
	log *slog.Logger,
	secretKey string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) (*AuthUsecase, error) {
	const op = "auth_usecase.NewAuthUsecase"
	log = log.With(slog.String("op", op))

	ctx := context.Background()
	userRepo, err := postgres.NewUserRepository(ctx, &cfg.Database, log)
	if err != nil {
		log.Error("failed to create user repository", slog.Any("error", err))
		return nil, err
	}

	log.Info("auth usecase initialized successfully")
	return &AuthUsecase{
		cfg:           cfg,
		log:           log,
		userRepo:      userRepo,
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, user *models.User) (string, error) {
	const op = "auth_usecase.Register"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("email", user.Email),
	)

	log.Info("attempting to register user")

	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, e.ErrUserNotFound) {
		log.Error("failed to check user existence", slog.Any("error", err))
		return "", err
	}

	if existingUser != nil {
		log.Warn("user already exists")
		return "", e.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.Any("error", err))
		return "", err
	}
	user.Password = string(hashedPassword)

	userID, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		return "", err
	}

	log.Info("user registered successfully", slog.String("user_id", userID))
	return userID, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (*models.TokenDetails, error) {
	const op = "auth_usecase.Login"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting user login")

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, e.ErrUserNotFound) {
			log.Warn("user not found")
		} else {
			log.Error("failed to get user by email", slog.Any("error", err))
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn("invalid credentials provided")
		return nil, e.ErrInvalidCredentials
	}

	tokens, err := uc.generateTokens(user.ID)
	if err != nil {
		log.Error("failed to generate tokens", slog.Any("error", err))
		return nil, err
	}

	log.Info("user logged in successfully", slog.String("user_id", user.ID))
	return tokens, nil
}

func (uc *AuthUsecase) ValidateToken(tokenString string) (string, error) {
	const op = "auth_usecase.ValidateToken"
	log := uc.log.With(slog.String("op", op))

	log.Debug("validating token")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Warn("invalid token signing method")
			return nil, e.ErrInvalidToken
		}
		return []byte(uc.secretKey), nil
	})

	if err != nil {
		log.Warn("failed to parse token", slog.Any("error", err))
		return "", e.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Warn("invalid token claims")
			return "", e.ErrInvalidToken
		}

		log.Debug("token validated successfully", slog.String("user_id", userID))
		return userID, nil
	}

	log.Warn("invalid token")
	return "", e.ErrInvalidToken
}

func (uc *AuthUsecase) RefreshToken(refreshToken string) (*models.TokenDetails, error) {
	const op = "auth_usecase.RefreshToken"
	log := uc.log.With(slog.String("op", op))

	log.Info("attempting to refresh token")

	userID, err := uc.ValidateToken(refreshToken)
	if err != nil {
		log.Warn("invalid refresh token", slog.Any("error", err))
		return nil, err
	}

	tokens, err := uc.generateTokens(userID)
	if err != nil {
		log.Error("failed to generate new tokens", slog.Any("error", err))
		return nil, err
	}

	log.Info("token refreshed successfully", slog.String("user_id", userID))
	return tokens, nil
}

func (uc *AuthUsecase) generateTokens(userID string) (*models.TokenDetails, error) {
	const op = "auth_usecase.generateTokens"
	log := uc.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	log.Debug("generating new tokens")

	now := time.Now()
	accessExpire := now.Add(uc.accessExpiry).Unix()
	refreshExpire := now.Add(uc.refreshExpiry).Unix()

	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     accessExpire,
		"iat":     now.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(uc.secretKey))
	if err != nil {
		log.Error("failed to sign access token", slog.Any("error", err))
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     refreshExpire,
		"iat":     now.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(uc.secretKey))
	if err != nil {
		log.Error("failed to sign refresh token", slog.Any("error", err))
		return nil, err
	}

	log.Debug("tokens generated successfully",
		slog.Int64("access_expire", accessExpire),
		slog.Int64("refresh_expire", refreshExpire),
	)

	return &models.TokenDetails{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		AtExpires:    accessExpire,
		RtExpires:    refreshExpire,
	}, nil
}
