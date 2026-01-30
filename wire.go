
package main

import (
	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/handler"
	"github.com/gadhittana01/cases-app-server/routes"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/google/wire"
)

func InitializeApp(db utils.PGXPool, config *utils.Config) (*App, error) {
	wire.Build(
		repository.NewRepository,
		NewS3Client,
		NewPresignClient,
		NewPusherClient,
		service.NewUserService,
		service.NewCaseService,
		service.NewQuoteService,
		service.NewMarketplaceService,
		service.NewPaymentService,
		service.NewFileService,
		handler.NewUserHandler,
		handler.NewCaseHandler,
		handler.NewQuoteHandler,
		handler.NewMarketplaceHandler,
		handler.NewPaymentHandler,
		handler.NewFileHandler,
		handler.NewWebhookHandler,
		routes.SetupRoutes,
		NewApp,
	)
	return &App{}, nil
}
