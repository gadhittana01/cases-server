package main

import (
	"github.com/gadhittana01/cases-app-server/providers"
	"github.com/gadhittana01/cases-modules/utils"
)

// Re-export provider functions for wire
var (
	NewS3Client      = providers.NewS3Client
	NewPresignClient = providers.NewPresignClient
	NewPusherClient  = providers.NewPusherClient
)

func main() {
	config := utils.CheckAndSetConfig("./config", "app")
	DBpool, err := utils.ConnectDBPool(config.DBConnString)
	if err != nil {
		panic(err)
	}
	DB, err := utils.ConnectDB(config.DBConnString)
	if err != nil {
		panic(err)
	}

	baseConfig := &utils.BaseConfig{
		MigrationURL: config.MigrationURL,
		DBName:       config.DBName,
	}

	if err := utils.RunMigrationPool(DB, baseConfig); err != nil {
		panic(err)
	}

	// Initialize app using wire-generated code
	appInstance, err := InitializeApp(DBpool, config)
	if err != nil {
		panic(err)
	}

	appInstance.Start()
}
