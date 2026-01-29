package api

import (
	"net/http"
	"sync"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/handler"
	"github.com/gadhittana01/cases-app-server/providers"
	"github.com/gadhittana01/cases-app-server/routes"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/gin-gonic/gin"
)

var (
	router   *gin.Engine
	initOnce sync.Once
)

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(initApp)

	if router != nil {
		router.ServeHTTP(w, r)
	} else {
		http.Error(w, "Application not initialized", http.StatusInternalServerError)
	}
}

func initApp() {
	// Use existing config loading pattern from main.go
	config := utils.CheckAndSetConfig("", "")

	// Use existing database connection pattern from main.go
	DBpool, err := utils.ConnectDBPool(config.DBConnString)
	if err != nil {
		return
	}

	DB, err := utils.ConnectDB(config.DBConnString)
	if err != nil {
		return
	}

	// Use existing migration pattern from main.go
	baseConfig := &utils.BaseConfig{
		MigrationURL: config.MigrationURL,
		DBName:       config.DBName,
	}

	if err := utils.RunMigrationPool(DB, baseConfig); err != nil {
		return
	}

	// Use InitializeApp from wire_gen.go - exact same code as wire_gen.go line 19-42
	repositoryRepository := repository.NewRepository(DBpool)
	userService := service.NewUserService(repositoryRepository, config)
	userHandler := handler.NewUserHandler(userService)
	caseService := service.NewCaseService(repositoryRepository)
	client, err := providers.NewS3Client(config)
	if err != nil {
		return
	}
	presignClient := providers.NewPresignClient(client)
	fileService := service.NewFileService(repositoryRepository, client, presignClient, config)
	caseHandler := handler.NewCaseHandler(caseService, fileService)
	quoteService := service.NewQuoteService(repositoryRepository)
	quoteHandler := handler.NewQuoteHandler(quoteService)
	marketplaceService := service.NewMarketplaceService(repositoryRepository)
	marketplaceHandler := handler.NewMarketplaceHandler(marketplaceService)
	pusherClient := providers.NewPusherClient(config)
	paymentService := service.NewPaymentService(repositoryRepository, config, pusherClient)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	fileHandler := handler.NewFileHandler(fileService)
	webhookHandler := handler.NewWebhookHandler(paymentService, config)
	engine := routes.SetupRoutes(userHandler, caseHandler, quoteHandler, marketplaceHandler, paymentHandler, fileHandler, webhookHandler, config)
	router = engine
}
