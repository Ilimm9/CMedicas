package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Ilimm9/CMedicas/Repositories"
	"github.com/Ilimm9/CMedicas/Respuestas"
	"github.com/Ilimm9/CMedicas/dto"
	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Crear una nueva persona
func PostPersona(c *gin.Context) {
	var input dto.PersonaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	fecha, err := time.Parse("2006-01-02", input.FechaNacimiento)
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "Formato de fecha inválido (usa YYYY-MM-DD)")
		return
	}

	persona := models.Persona{
		Nombre:          input.Nombre,
		ApellidoPaterno: input.ApellidoPaterno,
		ApellidoMaterno: input.ApellidoMaterno,
		Telefono:        input.Telefono,
		FechaNacimiento: fecha,
		Genero:          input.Genero,
		Direccion:       input.Direccion,
	}

	if err := repositories.CrearPersona(&persona); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := dto.PersonaResponse{
		ID:              persona.ID,
		Nombre:          persona.Nombre,
		ApellidoPaterno: persona.ApellidoPaterno,
		ApellidoMaterno: persona.ApellidoMaterno,
		Telefono:        persona.Telefono,
		FechaNacimiento: persona.FechaNacimiento.Format("2006-01-02"),
		Genero:          persona.Genero,
		Direccion:       persona.Direccion,
	}

	respuestas.RespondSuccess(c, http.StatusCreated, response)
}


// Obtener una persona por ID
func GetPersona(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	persona, err := repositories.ObtenerPersonaPorID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuestas.RespondError(c, http.StatusNotFound, "Persona no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := dto.PersonaResponse{
		ID:              persona.ID,
		Nombre:          persona.Nombre,
		ApellidoPaterno: persona.ApellidoPaterno,
		ApellidoMaterno: persona.ApellidoMaterno,
		Telefono:        persona.Telefono,
		FechaNacimiento: persona.FechaNacimiento.Format("2006-01-02"),
		Genero:          persona.Genero,
		Direccion:       persona.Direccion,
	}

	respuestas.RespondSuccess(c, http.StatusOK, response)
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

	var input dto.PersonaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	fecha, err := time.Parse("2006-01-02", input.FechaNacimiento)
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "Formato de fecha inválido (usa YYYY-MM-DD)")
		return
	}

	persona := models.Persona{
		Nombre:          input.Nombre,
		ApellidoPaterno: input.ApellidoPaterno,
		ApellidoMaterno: input.ApellidoMaterno,
		Telefono:        input.Telefono,
		FechaNacimiento: fecha,
		Genero:          input.Genero,
		Direccion:       input.Direccion,
	}

	if err := repositories.ActualizarPersona(uint(id), &persona); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, "Persona actualizada correctamente")
}


// Elimina una persona
func DeletePersona(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := repositories.EliminarPersona(uint(id)); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, "Persona eliminada correctamente")
}


