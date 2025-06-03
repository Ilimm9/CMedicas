package migrate

import (
	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/models"
)

func Migrations(){
	initializers.DB.AutoMigrate(&models.Persona{})
	initializers.DB.AutoMigrate(&models.Usuario{})
	initializers.DB.AutoMigrate(&models.Medico{})
	initializers.DB.AutoMigrate(&models.Cita{})
	initializers.DB.AutoMigrate(&models.Horario{})
	initializers.DB.AutoMigrate(&models.Notificacion{})
	initializers.DB.AutoMigrate(&models.Observacion{})
}