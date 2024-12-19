package webui

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
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

	command := fmt.Sprintf("AUTH %s %s", username, password)
	response, err := utils.SendCommand(command)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}

	if commandResponse.Status == "OK" {
		c.Cookie(&fiber.Cookie{
			Name:  "auth_token",
			Value: "valid_token",
		})
		c.Cookie(&fiber.Cookie{
			Name:  "username",
			Value: username,
		})
		return c.Redirect("/")
	} else {
		c.Status(fiber.StatusUnauthorized)
		return c.Render("login", fiber.Map{
			"Title":   "Login",
			"Message": commandResponse.Message,
		})
		// return c.Redirect("/login")
	}
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
