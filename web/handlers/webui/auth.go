package webui

import (
	"github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/gofiber/fiber/v2"
)

func LoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"Title":   "Login",
		"Message": "",
	})
}

func Authenticate(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	err := persistdb.OpenDB()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer persistdb.CloseDB()

	ok, err := persistdb.AuthenticateUser(username, password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}
	// get user
	persistedUser, err := persistdb.GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	userdto, err := persistdb.MaapUserToUserDTO(persistedUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	token, err := persistdb.GenerateJWTToken(userdto)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	c.Cookie(&fiber.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	return c.Redirect("/")
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:  "auth_token",
		Value: "",
	})
	c.Cookie(&fiber.Cookie{
		Name:  "username",
		Value: "",
	})
	return c.Redirect("/login")
}
