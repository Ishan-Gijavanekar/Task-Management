package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/testutil"
)

func newMemoryUserService() (*UserService, *testutil.MemoryUserRepository) {
	repo := testutil.NewMemoryUserRepository()
	var userRepo repository.UserRespository = repo
	return NewUserService(&userRepo), repo
}

func TestUserServiceCRUDAndPagination(t *testing.T) {
	service, _ := newMemoryUserService()
	ctx := context.Background()

	first, err := service.Create(ctx, CreateUserInput{Name: " Ada Lovelace ", Email: " ADA@example.com "})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if first.Name != "Ada Lovelace" || first.Email != "ada@example.com" {
		t.Fatalf("Create() did not normalize input: %#v", first)
	}
	if _, err := service.Create(ctx, CreateUserInput{Name: "Grace Hopper", Email: "grace@example.com"}); err != nil {
		t.Fatalf("second Create() error: %v", err)
	}

	page, err := service.List(ctx, ListUserInput{Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if page.Total != 2 || len(page.Users) != 1 {
		t.Fatalf("List() = total %d, users %d; want 2, 1", page.Total, len(page.Users))
	}

	updated, err := service.Update(ctx, first.ID.Hex(), UpdateUserInput{Name: "Ada Byron", Email: "ada.byron@example.com"})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if updated.Name != "Ada Byron" || updated.Role != domain.UserRoleUser {
		t.Fatalf("unexpected updated user: %#v", updated)
	}

	if err := service.Delete(ctx, first.ID.Hex()); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if _, err := service.GetByID(ctx, first.ID.Hex()); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrUserNotFound", err)
	}
}

func TestUserServiceRejectsInvalidInput(t *testing.T) {
	service, _ := newMemoryUserService()
	if _, err := service.Create(context.Background(), CreateUserInput{Name: "", Email: "invalid"}); !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("Create() error = %v, want ErrInvalidName", err)
	}
}
