package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Ilimm9/CMedicas/Repositories"
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
	var input models.Persona

	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "Datos inválidos: "+err.Error())
		return
	}

	if err := repositories.CrearPersona(&input); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al crear persona: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusCreated, input)
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

	persona, err := repositories.ObtenerPersonaPorID(id)
	if err != nil {
		respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		return
	}

	if err := c.ShouldBindJSON(persona); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "Datos inválidos: "+err.Error())
		return
	}

	if err := repositories.ActualizarPersona(persona); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al actualizar persona: "+err.Error())
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

	// Verificar si la persona existe
	_, err = repositories.ObtenerPersonaPorID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar persona: "+err.Error())
		}
		return
	}

	// Eliminar si existe
	if err := repositories.EliminarPersona(id); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al eliminar persona: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, gin.H{"mensaje": "Persona eliminada correctamente"})
}

