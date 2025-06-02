package main

import (
	"CMedicas/controllers"
	"CMedicas/initializers"
	"CMedicas/migrate"

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
	persona := r.Group("persona")
	{
		persona.POST("/agregar", controllers.PostPersona) 
		persona.GET("", controllers.GetAllPersonas)
		persona.GET("/:id", controllers.GetPersona)
		persona.PUT("/:id", controllers.UpdatePersona)
		persona.DELETE("/eliminar", controllers.DeletePersona)

	}



	r.Run()
}
