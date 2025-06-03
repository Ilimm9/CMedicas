package controllers

import (
	"net/http"
	"strconv"
	"time"

	"CMedicas/initializers"
	"CMedicas/models"
	"CMedicas/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ObservacionInput struct {
	CitaID        uint   `json:"cita_id" binding:"required"`
	Observaciones string `json:"observaciones" binding:"required"`
	Diagnostico   string `json:"diagnostico"`
}

// Crear  observación
func PostObservacion(c *gin.Context) {
	var input ObservacionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verificar que la cita existe
	var cita models.Cita
	if err := initializers.GetDB().First(&cita, input.CitaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusBadRequest, "Cita no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al verificar cita: "+err.Error())
		}
		return
	}

	// Verificar que la cita esté en estado "completada"
	if cita.Estado != "completada" {
		utils.RespondError(c, http.StatusBadRequest, "Solo se pueden agregar observaciones a citas completadas")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	observacion := models.Observacion{
		CitaID:        input.CitaID,
		Observaciones: input.Observaciones,
		Diagnostico:   input.Diagnostico,
		FechaRegistro: time.Now(),
	}

	if err := tx.Create(&observacion).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al guardar observación: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	if err := initializers.GetDB().
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Paciente.Persona").
		Preload("Cita.Medico").
		Preload("Cita.Medico.Usuario").
		Preload("Cita.Medico.Usuario.Persona").
		First(&observacion, observacion.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos de la observación: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, observacion)
}

// Obtener una observación por ID
func GetObservacion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var observacion models.Observacion
	result := initializers.GetDB().
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Paciente.Persona").
		Preload("Cita.Medico").
		Preload("Cita.Medico.Usuario").
		Preload("Cita.Medico.Usuario.Persona").
		First(&observacion, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Observación no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar observación: "+result.Error.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, observacion)
}

// Observación de una cita específica
func GetObservacionPorCita(c *gin.Context) {
	citaID, err := strconv.Atoi(c.Param("cita_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID de cita inválido")
		return
	}

	var observacion models.Observacion
	result := initializers.GetDB().
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Paciente.Persona").
		Preload("Cita.Medico").
		Preload("Cita.Medico.Usuario").
		Preload("Cita.Medico.Usuario.Persona").
		Where("cita_id = ?", citaID).
		First(&observacion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "No se encontró observación para esta cita")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar observación: "+result.Error.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, observacion)
}

// Actualizar una observación 
func UpdateObservacion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input struct {
		Observaciones string `json:"observaciones"`
		Diagnostico   string `json:"diagnostico"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	var observacion models.Observacion
	if err := tx.First(&observacion, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Observación no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar observación: "+err.Error())
		}
		return
	}

	// Actualizar solo los campos proporcionados
	if input.Observaciones != "" {
		observacion.Observaciones = input.Observaciones
	}
	if input.Diagnostico != "" {
		observacion.Diagnostico = input.Diagnostico
	}

	if err := tx.Save(&observacion).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al actualizar observación: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos actualizados para la respuesta
	if err := initializers.GetDB().
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Paciente.Persona").
		Preload("Cita.Medico").
		Preload("Cita.Medico.Usuario").
		Preload("Cita.Medico.Usuario.Persona").
		First(&observacion, observacion.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, observacion)
}

// Eliminar una observación
func DeleteObservacion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	result := tx.Delete(&models.Observacion{}, id)
	if result.Error != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al eliminar observación: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusNotFound, "Observación no encontrada")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "Observación eliminada correctamente"})
}