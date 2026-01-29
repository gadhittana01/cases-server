package main

import (
	"github.com/gadhittana01/go-modules-dependencies/utils"
	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
	config *utils.Config
}

func (a *App) Start() {
	port := a.config.Port
	if port == "" {
		port = "8000"
	}
	a.router.Run(":" + port)
}

func NewApp(router *gin.Engine, config *utils.Config) *App {
	return &App{
		router: router,
		config: config,
	}
}
