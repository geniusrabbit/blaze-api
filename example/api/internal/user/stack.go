package userstack

import (
	"github.com/geniusrabbit/blaze-api/repository/user"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	useruc "github.com/geniusrabbit/blaze-api/repository/user/usecase"
)

// Repository composes core + email + password repositories.
type Repository[T user.AuthCapableModel] interface {
	user.Repository[T]
	user.EmailRepository[T]
	user.PasswordRepository[T]
}

type compositeRepository[T user.AuthCapableModel] struct {
	user.Repository[T]
	user.EmailRepository[T]
	user.PasswordRepository[T]
}

// NewRepository wires core, email, and password repositories for auth-capable user type.
func NewRepository[T user.AuthCapableModel](newModel func() T) Repository[T] {
	core := userrepo.NewRepository(newModel)
	return &compositeRepository[T]{
		Repository:         core,
		EmailRepository:    userrepo.NewEmailRepository(core, newModel),
		PasswordRepository: userrepo.NewPasswordRepository(core, newModel),
	}
}

// Module groups user repositories and usecases for GraphQL/auth wiring.
type Module[T user.AuthCapableModel] struct {
	NewModel func() T
	Repo     Repository[T]
	Core     user.Usecase[T]
	Email    user.EmailUsecase[T]
	Password user.PasswordUsecase[T]
}

// NewModule wires the full auth-capable user stack (example/api).
func NewModule[T user.AuthCapableModel](newModel func() T) Module[T] {
	repo := NewRepository(newModel)
	core := useruc.NewUsecase(repo, newModel)
	return Module[T]{
		NewModel: newModel,
		Repo:     repo,
		Core:     core,
		Email:    useruc.NewEmailUsecase(repo, newModel),
		Password: useruc.NewPasswordUsecase(core, repo),
	}
}
