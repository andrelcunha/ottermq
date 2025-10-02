package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(c *fiber.Ctx) error {
	// Bypass authentication for static files
	currentPath := c.Path()

	if strings.HasPrefix(currentPath, "/assets/") ||
		strings.HasPrefix(currentPath, "/icons/") ||
		strings.HasPrefix(currentPath, "/images/") ||
		currentPath == "/favicon.ico" ||
		currentPath == "/index.html" {
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
