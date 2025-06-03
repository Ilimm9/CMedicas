package routes

import (
	"CMedicas/controllers"
	"CMedicas/middlewares"

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


	// ================== RUTAS PROTEGIDAS (requieren autenticación) ==================
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Perfil de usuario
		protected.GET("/usuario/actual", controllers.GetCurrentUser)
		// protected.PUT("/usuario/actual", controllers.UpdateCurrentUser)

		// Personas (accesible para usuarios autenticados)
		persona := protected.Group("/personas")
		{
			persona.GET("", controllers.GetAllPersonas)
			persona.GET("/:id", controllers.GetPersona)
			persona.POST("", controllers.PostPersona) 
			persona.PUT("/:id", controllers.UpdatePersona)
		}

		// Médicos (accesible para usuarios autenticados)
		medico := protected.Group("/medicos")
		{
			medico.GET("", controllers.GetAllMedicos)
			medico.GET("/:id", controllers.GetMedico)
			medico.GET("/:id/horarios", controllers.GetHorariosPorMedico)
		}

		// Citas (accesible para pacientes y médicos)
		cita := protected.Group("/citas")
		{
			cita.POST("", controllers.PostCita)
			cita.GET("", controllers.GetCitasUsuarioActual) // Devuelve citas según rol
			cita.GET("/:id", controllers.GetCita)
			cita.PUT("/:id/cancelar", controllers.CancelarCita)
		}

		// Observaciones (accesible para médicos y pacientes)
		observacion := protected.Group("/observaciones")
		{
			observacion.GET("/cita/:cita_id", controllers.GetObservacionPorCita)
		}

		// Notificaciones
		// notificacion := protected.Group("/notificaciones")
		// {
		// 	notificacion.GET("", controllers.GetNotificacionesUsuarioActual)
		// 	notificacion.PUT("/:id/marcar-leida", controllers.MarcarNotificacionLeida)
		// }
	}

	// ================== RUTAS DE ADMINISTRADOR ==================
	admin := r.Group("/api/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		// Gestión completa de personas
		admin.DELETE("/personas/:id", controllers.DeletePersona)

		// Gestión completa de médicos
		admin.POST("/medicos", controllers.PostMedico)
		admin.PUT("/medicos/:id", controllers.UpdateMedico)
		admin.DELETE("/medicos/:id", controllers.DeleteMedico)

		// Gestión de horarios médicos
		admin.POST("/medicos/:id/horarios", controllers.PostHorario)
		admin.PUT("/horarios/:id", controllers.UpdateHorario)
		admin.DELETE("/horarios/:id", controllers.DeleteHorario)

		// Gestión completa de citas
		admin.PUT("/citas/:id", controllers.UpdateCita)
		admin.DELETE("/citas/:id", controllers.DeleteCita)
		admin.GET("/citas/todas", controllers.GetAllCitas)

		// Gestión de observaciones
		admin.POST("/observaciones", controllers.PostObservacion)
		admin.PUT("/observaciones/:id", controllers.UpdateObservacion)
		admin.DELETE("/observaciones/:id", controllers.DeleteObservacion)

		// Gestión de notificaciones
		admin.POST("/notificaciones", controllers.PostNotificacion)
		// admin.GET("/notificaciones/todas", controllers.GetAllNotificaciones)
		admin.DELETE("/notificaciones/:id", controllers.DeleteNotificacion)

	}

}