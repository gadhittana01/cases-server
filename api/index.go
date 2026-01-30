package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gadhittana01/cases-app-server/db/repository"
	appHandler "github.com/gadhittana01/cases-app-server/handler"
	"github.com/gadhittana01/cases-app-server/providers"
	"github.com/gadhittana01/cases-app-server/routes"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/gin-gonic/gin"
)

var (
	router   *gin.Engine
	initOnce sync.Once
	initErr  error
)



func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handler called: %s %s", r.Method, r.URL.Path)

	initOnce.Do(initApp)

	if initErr != nil {
		log.Printf("Initialization error: %v", initErr)

		http.Error(w, "Application initialization failed: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	if router == nil {
		log.Println("Router is nil - initialization may have failed silently")
		http.Error(w, "Application not initialized", http.StatusInternalServerError)
		return
	}


	router.ServeHTTP(w, r)
}

func initApp() {
	log.Println("=== Starting application initialization ===")


	config := utils.CheckAndSetConfig("", "")

	log.Printf("Config loaded - DB_CONN_STRING present: %v", config.DBConnString != "")

	if config.DBConnString == "" {
		initErr = http.ErrMissingFile
		log.Println("ERROR: DB_CONN_STRING is not set in environment variables")
		return
	}


	log.Println("Connecting to database...")
	DBpool, err := utils.ConnectDBPool(config.DBConnString)
	if err != nil {
		initErr = err
		log.Printf("ERROR: Failed to connect to database: %v", err)
		return
	}
	log.Println("Database connection pool created successfully")

	DB, err := utils.ConnectDB(config.DBConnString)
	if err != nil {
		initErr = err
		log.Printf("ERROR: Failed to connect to database (sql.DB): %v", err)
		return
	}
	log.Println("Database connection (sql.DB) created successfully")


	baseConfig := &utils.BaseConfig{
		MigrationURL: config.MigrationURL,
		DBName:       config.DBName,
	}




	log.Println("Running migrations...")
	if err := utils.RunMigrationPool(DB, baseConfig); err != nil {
		log.Printf("WARNING: Migration error (continuing): %v", err)

	}


	log.Println("Initializing services and handlers...")
	repositoryRepository := repository.NewRepository(DBpool)
	userService := service.NewUserService(repositoryRepository, config)
	userHandler := appHandler.NewUserHandler(userService)
	caseService := service.NewCaseService(repositoryRepository)
	client, err := providers.NewS3Client(config)
	if err != nil {
		initErr = err
		log.Printf("ERROR: Failed to create S3 client: %v", err)
		return
	}
	presignClient := providers.NewPresignClient(client)
	fileService := service.NewFileService(repositoryRepository, client, presignClient, config)
	caseHandler := appHandler.NewCaseHandler(caseService, fileService)
	quoteService := service.NewQuoteService(repositoryRepository)
	quoteHandler := appHandler.NewQuoteHandler(quoteService)
	marketplaceService := service.NewMarketplaceService(repositoryRepository)
	marketplaceHandler := appHandler.NewMarketplaceHandler(marketplaceService)
	pusherClient := providers.NewPusherClient(config)
	paymentService := service.NewPaymentService(repositoryRepository, config, pusherClient)
	paymentHandler := appHandler.NewPaymentHandler(paymentService)
	fileHandler := appHandler.NewFileHandler(fileService)
	webhookHandler := appHandler.NewWebhookHandler(paymentService, config)

	engine := routes.SetupRoutes(userHandler, caseHandler, quoteHandler, marketplaceHandler, paymentHandler, fileHandler, webhookHandler, config)
	router = engine
}
