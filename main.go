package main

import (
	"CMedicas/initializers"
	"CMedicas/migrate"
	"CMedicas/routes"

	"github.com/gin-gonic/gin"
)


func init(){
	initializers.LoadEnv()
	initializers.ConnectDB()
	migrate.Migrations()
}

func main() {
	r := gin.Default()

	// Rutas
	routes.AdminRutas(r)

	r.Run()
}
