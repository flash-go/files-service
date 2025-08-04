package main

// @title		files-service
// @version		1.0
// @BasePath	/

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"log"
	"os"
	"strconv"

	// Framework
	//
	// Core of the Flash Framework. Contains the fundamental components of
	// the application.

	"github.com/flash-go/flash/http"
	"github.com/flash-go/flash/http/client"
	"github.com/flash-go/flash/http/server"

	// SDK
	//
	// A high-level software development toolkit based on the Flash Framework
	// for building highly efficient and fault-tolerant applications.

	"github.com/flash-go/sdk/config"
	"github.com/flash-go/sdk/errors"
	"github.com/flash-go/sdk/logger"
	"github.com/flash-go/sdk/services/users"
	"github.com/flash-go/sdk/state"
	"github.com/flash-go/sdk/telemetry"

	// Implementations

	//// Handlers
	httpDirsHandlerAdapterImpl "github.com/flash-go/files-service/internal/adapter/handler/dirs/http"
	httpFilesHandlerAdapterImpl "github.com/flash-go/files-service/internal/adapter/handler/files/http"

	//// Repository
	dirsRepositoryAdapterImpl "github.com/flash-go/files-service/internal/adapter/repository/dirs"
	filesRepositoryAdapterImpl "github.com/flash-go/files-service/internal/adapter/repository/files"

	//// Services
	dirsServiceImpl "github.com/flash-go/files-service/internal/service/dirs"
	filesServiceImpl "github.com/flash-go/files-service/internal/service/files"

	// Config
	internalConfig "github.com/flash-go/files-service/internal/config"

	// Other
	_ "github.com/flash-go/files-service/docs"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Create state service
	stateService := state.New(os.Getenv("CONSUL_AGENT"))

	// Create config
	cfg := config.New(
		stateService,
		os.Getenv("SERVICE_NAME"),
	)

	// Create logger service
	loggerService := logger.NewConsole()

	// Convert log level to int
	logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Fatalf("invalid log level")
	}

	// Set log level
	loggerService.SetLevel(logLevel)

	// Create telemetry service
	telemetryService := telemetry.NewGrpc(cfg)

	// Collect metrics
	telemetryService.CollectGoRuntimeMetrics(collectGoRuntimeMetricsTimeout)

	// Create http client
	httpClient := client.New()

	// Use state service
	httpClient.UseState(stateService)

	// Use telemetry service
	httpClient.UseTelemetry(telemetryService)

	// Create http server
	httpServer := server.New()

	// Use telemetry service
	httpServer.UseTelemetry(telemetryService)

	// Use logger service
	httpServer.UseLogger(loggerService)

	// Use state service
	httpServer.UseState(stateService)

	// Use Swagger
	httpServer.UseSwagger()

	// Set error response status map
	httpServer.SetErrorResponseStatusMap(
		&server.ErrorResponseStatusMap{
			errors.ErrBadRequest:   400,
			errors.ErrUnauthorized: 401,
			errors.ErrForbidden:    403,
			errors.ErrNotFound:     404,
		},
	)

	// Set max request body size
	httpServer.SetServerMaxRequestBodySize(1024 * 1024 * 1024 * 8) // 8GB

	// Create repository
	dirsRepository := dirsRepositoryAdapterImpl.New(
		&dirsRepositoryAdapterImpl.Config{
			StoreLocalRootPath: cfg.Get(internalConfig.StoreLocalRootPathOptKey),
		},
	)
	filesRepository := filesRepositoryAdapterImpl.New(
		&filesRepositoryAdapterImpl.Config{
			StoreLocalRootPath: cfg.Get(internalConfig.StoreLocalRootPathOptKey),
		},
	)

	// Create services
	dirsService := dirsServiceImpl.New(
		&dirsServiceImpl.Config{
			DirsRepository: dirsRepository,
		},
	)
	filesService := filesServiceImpl.New(
		&filesServiceImpl.Config{
			FilesRepository: filesRepository,
		},
	)

	// Create handlers
	dirsHandler := httpDirsHandlerAdapterImpl.New(
		&httpDirsHandlerAdapterImpl.Config{
			DirsService: dirsService,
		},
	)
	filesHandler := httpFilesHandlerAdapterImpl.New(
		&httpFilesHandlerAdapterImpl.Config{
			FilesService: filesService,
		},
	)

	// Create users middleware
	usersMiddleware := users.NewMiddleware(
		&users.MiddlewareConfig{
			UsersService: cfg.Get(internalConfig.UsersServiceNameOptKey),
			HttpClient:   httpClient,
		},
	)

	// Get admin role
	adminRole := cfg.Get(internalConfig.UsersAdminRoleOptKey)

	// Add routes
	httpServer.
		// Dirs

		// Create dir (admin)
		AddRoute(
			http.MethodPost,
			"/admin/dirs",
			dirsHandler.AdminCreateDir,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).
		// Delete dir (admin)
		AddRoute(
			http.MethodDelete,
			"/admin/dirs",
			dirsHandler.AdminDeleteDir,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).
		// Rename dir (admin)
		AddRoute(
			http.MethodPatch,
			"/admin/dirs",
			dirsHandler.AdminRenameDir,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).

		// Files

		// Create file (admin)
		AddRoute(
			http.MethodPost,
			"/admin/files",
			filesHandler.AdminCreateFile,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).
		// Get files (admin)
		AddRoute(
			http.MethodPost,
			"/admin/files/list",
			filesHandler.AdminListFiles,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).
		// Delete file (admin)
		AddRoute(
			http.MethodDelete,
			"/admin/files",
			filesHandler.AdminDeleteFile,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		).
		// Rename file (admin)
		AddRoute(
			http.MethodPatch,
			"/admin/files",
			filesHandler.AdminRenameFile,
			usersMiddleware.Auth(
				users.WithAuthRolesOption(adminRole),
			),
		)

	// Convert service port to int
	servicePort, err := strconv.Atoi(os.Getenv("SERVICE_PORT"))
	if err != nil || servicePort <= 0 {
		log.Fatalf("invalid service port")
	}

	// Register service
	if err := httpServer.RegisterService(
		os.Getenv("SERVICE_NAME"),
		os.Getenv("SERVICE_HOST"),
		servicePort,
	); err != nil {
		loggerService.Log().Err(err).Send()
	}

	// Convert server port to int
	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil || serverPort <= 0 {
		log.Fatal("invalid server port")
	}

	// Listen http server
	if err := <-httpServer.Listen(
		os.Getenv("SERVER_HOST"),
		serverPort,
	); err != nil {
		loggerService.Log().Err(err).Send()
	}
}
