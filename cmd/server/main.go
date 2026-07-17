package main

import (
	"gophkeeper/internal/app"
	"gophkeeper/internal/config"
)

func main() {
	config := config.InitConfiguration()

	app.Run(config)
}
