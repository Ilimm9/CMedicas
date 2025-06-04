package repositories

import (
	"errors"
	"gorm.io/gorm"
	"github.com/Ilimm9/CMedicas/models"
	"github.com/Ilimm9/CMedicas/initializers"
)

func ExistePersonaPorID(personaID uint) (bool, error) {
	var persona models.Persona
	err := initializers.GetDB().First(&persona, personaID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ExisteUsuarioPorCorreo(correo string) (bool, error) {
	var usuario models.Usuario
	err := initializers.GetDB().Where("correo = ?", correo).First(&usuario).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Buscar una persona por ID
func ObtenerPersonaPorID(id int) (*models.Persona, error) {
	var persona models.Persona
	if err := initializers.GetDB().First(&persona, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &persona, nil
}

// Crear una nueva persona
func CrearPersona(persona *models.Persona) error {
	if err := initializers.GetDB().Create(persona).Error; err != nil {
		return err
	}
	return nil
}

// Actualizar los datos de una persona existente
func ActualizarPersona(persona *models.Persona) error {
	if err := initializers.GetDB().Save(persona).Error; err != nil {
		return err
	}
	return nil
}

// Eliminar persona por ID
func EliminarPersona(id int) error {
	if err := initializers.GetDB().Delete(&models.Persona{}, id).Error; err != nil {
		return err
	}
	return nil
}
