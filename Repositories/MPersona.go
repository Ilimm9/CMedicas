package repositories

import (
	"github.com/Ilimm9/CMedicas/models"
	"github.com/Ilimm9/CMedicas/initializers"
)

func ExistePersonaPorID(id uint) bool {
	var count int64
	initializers.GetDB().Model(&models.Persona{}).Where("id = ?", id).Count(&count)
	return count > 0
}

// Buscar una persona por ID
func ObtenerPersonaPorID(id uint) (*models.Persona, error) {
	var persona models.Persona
	err := initializers.GetDB().First(&persona, id).Error
	return &persona, err
}

// Crear una nueva persona
func CrearPersona(persona *models.Persona) error {
	return initializers.GetDB().Create(persona).Error
}

// Actualizar los datos de una persona existente
func ActualizarPersona(id uint, persona *models.Persona) error {
	var existente models.Persona
	if err := initializers.GetDB().First(&existente, id).Error; err != nil {
		return err
	}

	persona.ID = existente.ID
	return initializers.GetDB().Save(persona).Error
}

// Eliminar persona por ID
func EliminarPersona(id uint) error {
	return initializers.GetDB().Delete(&models.Persona{}, id).Error
}