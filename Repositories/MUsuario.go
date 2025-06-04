package repositories

import (
	"github.com/Ilimm9/CMedicas/models"
	"github.com/Ilimm9/CMedicas/initializers"
)

// Buscar un usuario por su ID
func ObtenerUsuarioPorID(id uint) (*models.Usuario, error) {
	var usuario models.Usuario
	err := initializers.GetDB().First(&usuario, id).Error
	return &usuario, err
}


// Verificar si ya existe un usuario con el correo dado
func ExisteUsuarioPorCorreo(correo string) (bool, error) {
	var count int64
	err := initializers.GetDB().Model(&models.Usuario{}).Where("correo = ?", correo).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

//Guardar un nuevo usuario
func CrearUsuario(usuario *models.Usuario) error {
	return initializers.GetDB().Create(usuario).Error
}

func ExisteCorreo(correo string) bool {
	var count int64
	initializers.GetDB().Model(&models.Usuario{}).Where("correo = ?", correo).Count(&count)
	return count > 0
}


func ActualizarUsuario(id uint, usuario *models.Usuario) error {
	var existente models.Usuario
	if err := initializers.GetDB().First(&existente, id).Error; err != nil {
		return err
	}

	usuario.ID = existente.ID
	return initializers.GetDB().Save(usuario).Error
}


// Eliminar usuario por ID
func EliminarUsuario(id uint) error {
	return initializers.GetDB().Delete(&models.Usuario{}, id).Error
}

