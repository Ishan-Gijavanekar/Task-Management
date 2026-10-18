package http

import (
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: *userService,
	}
}

type UserResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
	Role      domain.UserRole   `json:"role"`
	Status    domain.UserStatus `json:"status"`
	Version   int64             `json:"version"`
}

type UpdateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ListUserResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Page  int64          `json:"page"`
	Limit int64          `json:"limit"`
}

func userToHttpResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		Version:   user.Version,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (u *UserHandler) GetUser(c *fiber.Ctx) error {
	userId := c.Params("id")
	ctx := c.UserContext()

	_, err := auth.RequireSelfOrAdmin(ctx, userId)
	if err != nil {
		return writeError(c, err)
	}

	user, err := u.userService.GetByID(ctx, userId)
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(userToHttpResponse(user))
}

func (u *UserHandler) ListUsers(c *fiber.Ctx) error {
	ctx := c.UserContext()

	_, err := auth.RequireRole(ctx, domain.UserRoleAdmin)
	if err != nil {
		return writeError(c, err)
	}

	page := int64(c.QueryInt("page", 0))
	limit := int64(c.QueryInt("limit", 10))

	result, err := u.userService.List(ctx, service.ListUserInput{
		Limit: limit,
		Page:  page,
	})
	if err != nil {
		return writeError(c, err)
	}

	users := make([]UserResponse, 0, len(result.Users))
	for _, user := range result.Users {
		users = append(users, userToHttpResponse(user))
	}

	return c.Status(fiber.StatusOK).JSON(
		ListUserResponse{
			Users: users,
			Page:  result.Page,
			Limit: result.Limit,
			Total: result.Total,
		},
	)
}

func (u *UserHandler) UpdateUsers(c *fiber.Ctx) error {
	var updateRequest UpdateRequest
	userId := c.Params("id")
	ctx := c.UserContext()

	_, err := auth.RequireSelfOrAdmin(ctx, userId)
	if err != nil {
		return writeError(c, err)
	}

	if err := c.BodyParser(&updateRequest); err != nil {
		return writeError(c, err)
	}

	user, err := u.userService.Update(ctx, userId, service.UpdateUserInput(updateRequest))
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(userToHttpResponse(user))
}

func (u *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userId := c.Params("id")
	ctx := c.UserContext()

	_, err := auth.RequireSelfOrAdmin(ctx, userId)
	if err != nil {
		return writeError(c, err)
	}

	err = u.userService.Delete(ctx, userId)
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User Deleted successfully"})
}
