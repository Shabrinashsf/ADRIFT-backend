package config

import (
	"ADRIFT-backend/internal/api/controller"
	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/pkg/mailer"
	"ADRIFT-backend/internal/pkg/storage"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/samber/do"
	"gorm.io/gorm"
)

// Provider is the dependency injection container
type Provider struct {
	injector *do.Injector
}

// NewProvider creates a new DI provider with database connection
func NewProvider(db *gorm.DB) *Provider {
	injector := do.New()

	// Register database singleton
	do.Provide(injector, func(i *do.Injector) (*gorm.DB, error) {
		return db, nil
	})

	// =========== SERVICES ===========
	// JWT Service
	do.Provide(injector, func(i *do.Injector) (service.JWTService, error) {
		return service.NewJWTService(), nil
	})

	// Mailer Service
	do.Provide(injector, func(i *do.Injector) (mailer.Mailer, error) {
		return mailer.NewMailer(), nil
	})

	// Request Validator
	do.Provide(injector, func(i *do.Injector) (*validate.Validator, error) {
		return validate.NewValidator(), nil
	})

	// File Storage Service
	do.Provide(injector, func(i *do.Injector) (storage.FileSystemStorage, error) {
		return storage.NewFileSystemStorage(), nil
	})

	// =========== REPOSITORIES ===========
	// User Repository
	do.Provide(injector, func(i *do.Injector) (repository.UserRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return repository.NewUserController(db), nil
	})

	// FRS Repository
	do.Provide(injector, func(i *do.Injector) (repository.FRSRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return repository.NewFRSRepository(db), nil
	})

	// =========== SERVICES ===========
	// User Service
	do.Provide(injector, func(i *do.Injector) (service.UserService, error) {
		userRepo := do.MustInvoke[repository.UserRepository](i)
		jwtService := do.MustInvoke[service.JWTService](i)
		mailerService := do.MustInvoke[mailer.Mailer](i)
		db := do.MustInvoke[*gorm.DB](i)
		return service.NewUserService(userRepo, jwtService, mailerService, db), nil
	})

	// FRS Service
	do.Provide(injector, func(i *do.Injector) (service.FRSService, error) {
		frsRepo := do.MustInvoke[repository.FRSRepository](i)
		storage := do.MustInvoke[storage.FileSystemStorage](i)
		db := do.MustInvoke[*gorm.DB](i)
		return service.NewFRSService(frsRepo, storage, db), nil
	})

	// =========== CONTROLLERS ===========
	// User Controller
	do.Provide(injector, func(i *do.Injector) (controller.UserController, error) {
		userService := do.MustInvoke[service.UserService](i)
		validator := do.MustInvoke[*validate.Validator](i)
		storage := do.MustInvoke[storage.FileSystemStorage](i)
		return controller.NewUserController(userService, validator, storage), nil
	})

	// FRS Controller
	do.Provide(injector, func(i *do.Injector) (controller.FRSController, error) {
		frsService := do.MustInvoke[service.FRSService](i)
		validator := do.MustInvoke[*validate.Validator](i)
		return controller.NewFRSController(frsService, validator), nil
	})

	do.Provide(injector, func(i *do.Injector) (controller.FileController, error) {
		storage := do.MustInvoke[storage.FileSystemStorage](i)
		return controller.NewFileController(storage), nil
	})

	return &Provider{
		injector: injector,
	}
}

// =========== INVOKE METHODS ===========

// InvokeJWTService returns the JWT service instance
func (p *Provider) InvokeJWTService() service.JWTService {
	return do.MustInvoke[service.JWTService](p.injector)
}

// InvokeMailerService returns the Mailer service instance
func (p *Provider) InvokeMailerService() mailer.Mailer {
	return do.MustInvoke[mailer.Mailer](p.injector)
}

// InvokeUserController returns the User controller instance
func (p *Provider) InvokeUserController() controller.UserController {
	return do.MustInvoke[controller.UserController](p.injector)
}

// InvokeFRSController returns the FRS controller instance
func (p *Provider) InvokeFRSController() controller.FRSController {
	return do.MustInvoke[controller.FRSController](p.injector)
}

// InvokeFileController returns the File controller instance
func (p *Provider) InvokeFileController() controller.FileController {
	return do.MustInvoke[controller.FileController](p.injector)
}

// InvokeUserService returns the User service instance
func (p *Provider) InvokeUserService() service.UserService {
	return do.MustInvoke[service.UserService](p.injector)
}

// InvokeUserRepository returns the User repository instance
func (p *Provider) InvokeUserRepository() repository.UserRepository {
	return do.MustInvoke[repository.UserRepository](p.injector)
}

// InvokeDatabase returns the database instance
func (p *Provider) InvokeDatabase() *gorm.DB {
	return do.MustInvoke[*gorm.DB](p.injector)
}

// Shutdown gracefully shuts down the provider and cleans up resources
func (p *Provider) Shutdown() {
	// The do library handles singleton lifecycle automatically
	// Additional cleanup can be added here if needed
}
