package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Ilimm9/CMedicas/Respuestas"
	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PersonaInput struct {
	Nombre          string    `json:"nombre" binding:"required"`
	ApellidoPaterno string    `json:"apellido_paterno" binding:"required"`
	ApellidoMaterno string    `json:"apellido_materno" binding:"required"`
	Telefono        string    `json:"telefono"`
	FechaNacimiento time.Time `json:"fecha_nacimiento"`
	Genero          string    `json:"genero" binding:"required,oneof=masculino femenino otro"`
	Direccion       string    `json:"direccion"`
}

// Crear una nueva persona
func PostPersona(c *gin.Context) {
	var input PersonaInput

	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	persona := models.Persona{
		Nombre:          input.Nombre,
		ApellidoPaterno: input.ApellidoPaterno,
		ApellidoMaterno: input.ApellidoMaterno,
		Telefono:        input.Telefono,
		FechaNacimiento: input.FechaNacimiento,
		Genero:          input.Genero,
		Direccion:       input.Direccion,
	}

	if err := tx.Create(&persona).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al guardar persona: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusCreated, persona)
}

// Obtener una persona por ID
func GetPersona(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var persona models.Persona
	result := initializers.GetDB().First(&persona, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar persona: "+result.Error.Error())
		}
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, persona)
}

// Obtener todas las personas
func GetAllPersonas(c *gin.Context) {
	var personas []models.Persona
	result := initializers.GetDB().Find(&personas)
	if result.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al obtener personas: "+result.Error.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, personas)
}

// Actualizar una persona existente
func UpdatePersona(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input PersonaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	var persona models.Persona
	if err := tx.First(&persona, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar persona: "+err.Error())
		}
		return
	}

	// Actualizar campos
	persona.Nombre = input.Nombre
	persona.ApellidoPaterno = input.ApellidoPaterno
	persona.ApellidoMaterno = input.ApellidoMaterno
	persona.Telefono = input.Telefono
	persona.FechaNacimiento = input.FechaNacimiento
	persona.Genero = input.Genero
	persona.Direccion = input.Direccion

	if err := tx.Save(&persona).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al actualizar persona: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, persona)
}

// Elimina una persona
func DeletePersona(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	result := tx.Delete(&models.Persona{}, id)
	if result.Error != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al eliminar persona: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, gin.H{"message": "Persona eliminada correctamente"})
}
