package auth

import (
	"net/http"
	"strings"
	"time"

	"iot-platform/pkg/config"
	"iot-platform/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:          cfg.Secret,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *JWTManager) GenerateAccessToken(userID, username string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "iot-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// GenerateRefreshToken 生成刷新令牌
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "iot-platform",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// ParseToken 解析令牌
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(errors.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(errors.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			c.Error(errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// RequireRoles 角色鉴权中间件
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.Error(errors.ErrForbidden)
			c.Abort()
			return
		}

		roleList, ok := userRoles.([]string)
		if !ok {
			c.Error(errors.ErrForbidden)
			c.Abort()
			return
		}

		roleMap := make(map[string]bool)
		for _, r := range roleList {
			roleMap[r] = true
		}

		for _, required := range roles {
			if roleMap[required] {
				c.Next()
				return
			}
		}

		c.Error(errors.ErrForbidden)
		c.Abort()
	}
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
