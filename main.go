package main

import (
	"log"

	"vvinamp/controller"
	"vvinamp/module"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// Register function register all API controllers to Mux
func Register(target controller.Target) {
	for _, controller := range target.Controllers {
		controller.Register(target.Gin)
	}
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	app := fx.New(
		module.Module,
		fx.Invoke(Register),
	)

	app.Run()
}
