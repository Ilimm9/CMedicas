package routes

import (
	"CMedicas/controllers"

	"github.com/gin-gonic/gin"
)

func AdminRutas(r *gin.Engine) {

	public := r.Group("/api")
	{
		// Autenticación
		public.POST("/auth/registro", controllers.RegistroCompleto)
		public.POST("/auth/login", controllers.Login)
		
		// public.GET("/medicos/disponibles", controllers.GetMedicosDisponibles)
		// public.GET("/especialidades", controllers.GetEspecialidades)
	}

}