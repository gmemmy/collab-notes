// Package middleware provides HTTP middleware components for authentication and authorization
// used throughout the application to secure API endpoints
package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected returns a middleware that validates JWT tokens and injects user ID into the request context.
// This middleware should be used on routes that require authentication.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid Authorization header"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		secret := os.Getenv("JWT_SECRET")
		tokenString = strings.TrimSpace(tokenString)

		token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["user-id"] == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		// Inject user ID into context
		c.Locals("user-id", claims["user-id"])

		return c.Next()
	}
}
