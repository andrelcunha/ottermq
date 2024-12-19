package api_admin

import (
	"github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/gofiber/fiber/v2"
)

// AddUser godoc
// @Summary Add a user
// @Description Add a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User details"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/admin/users [post]
func AddUser(c *fiber.Ctx) error {
	var user persistdb.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	err := persistdb.OpenDB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer persistdb.CloseDB()
	err = persistdb.AddUser(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User added successfully",
	})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} []User
// @Failure 500 {object} fiber.Map
// @Router /api/admin/users [get]
func GetUsers(c *fiber.Ctx) error {
	err := persistdb.OpenDB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer persistdb.CloseDB()
	users, err := persistdb.GetUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(users)
}

// Login godoc
// @Summary Login
// @Description Login
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User details"
// @Success 200 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/admin/login [post]
func Login(c *fiber.Ctx) error {
	var user persistdb.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	err := persistdb.OpenDB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer persistdb.CloseDB()

	ok, err := persistdb.AuthenticateUser(user.Username, user.Password)
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
	userdto, err := persistdb.GetUserByUsername(user.Username)
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})

}
