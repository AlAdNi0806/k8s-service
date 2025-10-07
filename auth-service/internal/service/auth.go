// internal/service/auth.go
package service

import (
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	redis    *redis.Client
}

func NewAuthService(userRepo *repository.UserRepository, redis *redis.Client) *AuthService {
	return &AuthService{userRepo: userRepo, redis: redis}
}

// Register — регистрирует пользователя
func (s *AuthService) Register(ctx context.Context, email, password string) error {
	logger := utils.NewHelperLogger("auth-service.service.register")
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.userRepo.Create(email, string(hashed))

	if err != nil {
		logger.LogError(ctx, "User not found during login attempt", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
		)
		return err
	}

	logger.LogInfo(ctx, "User not found during login attempt",
		log.KeyValue{Key: "email", Value: log.StringValue(email)},
	)

	return err
}

// Login — аутентифицирует и возвращает JWT-токен
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	logger := utils.NewHelperLogger("auth-service.service.login")

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		logger.LogError(ctx, "User not found during login attempt", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
		)
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.LogError(ctx, "User not found during login attempt", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
			log.KeyValue{Key: "user.id", Value: log.Int64Value(user.ID)},
			log.KeyValue{Key: "error", Value: log.StringValue(err.Error())},
		)
		return "", err
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		logger.LogError(ctx, "Failed to generate token", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
		)
		return "", err
	}

	// Генерируем уникальный ключ для хранения в Redis
	tokenKey := "token:" + generateTokenID()
	err = s.redis.Set(ctx, tokenKey, user.ID, 24*time.Hour).Err()
	if err != nil {
		logger.LogError(ctx, "Could not store token in Redis", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
		)
		return "", err
	}

	// Сохраняем связь токен → ключ (для отзыва)
	err = s.redis.Set(ctx, "user_token:"+token, tokenKey, 24*time.Hour).Err()
	if err != nil {
		logger.LogError(ctx, "Could not store token in Redis _2", err,
			log.KeyValue{Key: "email", Value: log.StringValue(email)},
		)
		// Опционально: удалить tokenKey при ошибке
		s.redis.Del(ctx, tokenKey)
		return "", err
	}

	return token, nil
}

// ValidateToken — проверяет валидность токена через JWT + Redis
func (s *AuthService) ValidateToken(ctx context.Context, token string) (int64, error) {
	userID, err := utils.ValidateToken(token)
	if err != nil {
		return 0, err
	}

	tokenKey, err := s.redis.Get(ctx, "user_token:"+token).Result()
	if err != nil {
		return 0, err // токен не найден → недействителен
	}

	storedUserID, err := s.redis.Get(ctx, tokenKey).Int64()
	if err != nil {
		return 0, err
	}
	if storedUserID != userID {
		return 0, err
	}

	return userID, nil
}

func generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
