package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(c *fiber.Ctx) error {
	// Bypass authentication for static files
	currentPath := c.Path()

	if strings.HasPrefix(currentPath, "/static/") ||
		strings.HasPrefix(currentPath, "/css/") ||
		strings.HasPrefix(currentPath, "/js/") ||
		currentPath == "/favicon.ico" ||
		currentPath == "/login" {
		log.Println("Bypassing authentication for", currentPath)
		return c.Next()
	}

	// arraay of paths that don't require authentication

	if !isAuthenticated(c) {
		log.Println("Unauthorized access to", currentPath)
		return c.Redirect("/login")
	} else {
		log.Println("Authenticated access to", currentPath)
	}
	return c.Next()
}

func isAuthenticated(c *fiber.Ctx) bool {
	token := c.Cookies("auth_token")
	return token != ""
}
