package main

import (
	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/migrate"
	"github.com/Ilimm9/CMedicas/routes"

	"github.com/gin-gonic/gin"
)


func init(){
	// solo para modo local
	// initializers.LoadEnv()
	initializers.ConnectDB()
	migrate.Migrations()
}

func main() {
	r := gin.Default()

	// Rutas
	routes.AdminRutas(r)

	r.Run()
}
