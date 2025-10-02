package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(c *fiber.Ctx) error {
	// Bypass authentication for static files
	currentPath := c.Path()

	if strings.HasPrefix(currentPath, "/static/") ||
		strings.HasPrefix(currentPath, "/css/") ||
		strings.HasPrefix(currentPath, "/js/") ||
		strings.HasPrefix(currentPath, "/images/") ||
		currentPath == "/favicon.ico" ||
		// strings.HasPrefix(currentPath, "/api/") ||
		currentPath == "/login" {
		return c.Next()
	}

	if !isAuthenticated(c) {
		return c.Redirect("/login")
	}
	return c.Next()
}

func isAuthenticated(c *fiber.Ctx) bool {
	token := c.Cookies("auth_token")
	return token != ""
}
