package api

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/models"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

func Authenticate(c *fiber.Ctx) error {
	var request models.AuthRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("AUTH %s %s",
		request.Username,
		request.Password)
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

	if commandResponse.Status == "ERROR" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": commandResponse.Message,
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token": commandResponse.Data,
		})
	}

}
