package routes

import (
	"log"

	"github.com/gadhittana01/cases-app-server/handler"
	"github.com/gadhittana01/cases-modules/middleware"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	userHandler *handler.UserHandler,
	caseHandler *handler.CaseHandler,
	quoteHandler *handler.QuoteHandler,
	marketplaceHandler *handler.MarketplaceHandler,
	paymentHandler *handler.PaymentHandler,
	fileHandler *handler.FileHandler,
	webhookHandler *handler.WebhookHandler,
	config *utils.Config,
) *gin.Engine {
	jwtSecret := config.JWTSecret
	r := gin.Default()

	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		log.Println("Health check request received")
		c.JSON(200, gin.H{"status": "ok"})
	})

	public := r.Group("/api/v1")
	{
		public.POST("/auth/signup/client", userHandler.Signup)
		public.POST("/auth/signup/lawyer", userHandler.Signup)
		public.POST("/auth/login", userHandler.Login)
		public.POST("/webhooks/stripe", webhookHandler.HandleStripeWebhook)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		api.GET("/auth/profile", userHandler.GetProfile)

		client := api.Group("")
		client.Use(middleware.RequireRole("client"))
		{
			client.GET("/client/cases", caseHandler.GetMyCases)
			client.POST("/client/cases", caseHandler.CreateCase)
			client.GET("/client/cases/:id", caseHandler.GetCaseByID)
			client.POST("/client/cases/:id/files", caseHandler.UploadFile)
			client.POST("/client/quotes/accept", paymentHandler.AcceptQuote)
		}

		lawyer := api.Group("")
		lawyer.Use(middleware.RequireRole("lawyer"))
		{
			lawyer.GET("/lawyer/marketplace", marketplaceHandler.ListOpenCases)
			lawyer.GET("/lawyer/marketplace/cases/:id", marketplaceHandler.GetCaseForMarketplace)
			lawyer.GET("/lawyer/marketplace/cases/:id/quotes/my", quoteHandler.GetMyQuoteForCase)
			lawyer.POST("/lawyer/marketplace/cases/:id/quotes", quoteHandler.CreateQuote)
			lawyer.PUT("/lawyer/marketplace/cases/:id/quotes", quoteHandler.UpdateQuote)
			lawyer.GET("/lawyer/quotes", quoteHandler.GetMyQuotes)
		}

		api.GET("/files/:id/download", fileHandler.GenerateDownloadURL)
	}

	return r
}
