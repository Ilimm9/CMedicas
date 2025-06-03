package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/models"
	"github.com/Ilimm9/CMedicas/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HorarioInput struct {
	MedicoID   uint      `json:"medico_id" binding:"required"`
	DiaSemana  string    `json:"dia_semana" binding:"required,oneof=Lunes Martes Miércoles Jueves Viernes Sábado Domingo"`
	HoraInicio time.Time `json:"hora_inicio" binding:"required"`
	HoraFin    time.Time `json:"hora_fin" binding:"required"`
}

// PostHorario crea un nuevo horario
func PostHorario(c *gin.Context) {
	var input HorarioInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validar que el médico existe
	var medico models.Medico
	if err := initializers.GetDB().First(&medico, input.MedicoID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusBadRequest, "Médico no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al verificar médico: "+err.Error())
		}
		return
	}

	// Validar que la hora de fin sea mayor que la de inicio
	if input.HoraFin.Before(input.HoraInicio) || input.HoraFin.Equal(input.HoraInicio) {
		utils.RespondError(c, http.StatusBadRequest, "La hora de fin debe ser posterior a la hora de inicio")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	horario := models.Horario{
		MedicoID:   input.MedicoID,
		DiaSemana:  input.DiaSemana,
		HoraInicio: input.HoraInicio,
		HoraFin:    input.HoraFin,
	}

	if err := tx.Create(&horario).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al guardar horario: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos relacionados para la respuesta
	if err := initializers.GetDB().
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&horario, horario.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos del horario: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, horario)
}

// GetHorario obtiene un horario por ID
func GetHorario(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var horario models.Horario
	result := initializers.GetDB().
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&horario, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Horario no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar horario: "+result.Error.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, horario)
}

// GetAllHorarios obtiene todos los horarios
func GetAllHorarios(c *gin.Context) {
	var horarios []models.Horario
	result := initializers.GetDB().
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		Find(&horarios)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener horarios: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, horarios)
}

// GetHorariosPorMedico obtiene los horarios de un médico específico
func GetHorariosPorMedico(c *gin.Context) {
	medicoID, err := strconv.Atoi(c.Param("medico_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID de médico inválido")
		return
	}

	var horarios []models.Horario
	result := initializers.GetDB().
		Preload("Medico").
		Where("medico_id = ?", medicoID).
		Find(&horarios)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener horarios: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, horarios)
}

// UpdateHorario actualiza un horario existente
func UpdateHorario(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input struct {
		DiaSemana  string    `json:"dia_semana" binding:"omitempty,oneof=Lunes Martes Miércoles Jueves Viernes Sábado Domingo"`
		HoraInicio time.Time `json:"hora_inicio"`
		HoraFin    time.Time `json:"hora_fin"`
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

	var horario models.Horario
	if err := tx.First(&horario, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Horario no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar horario: "+err.Error())
		}
		return
	}

	// Actualizar solo los campos proporcionados
	if input.DiaSemana != "" {
		horario.DiaSemana = input.DiaSemana
	}
	if !input.HoraInicio.IsZero() {
		horario.HoraInicio = input.HoraInicio
	}
	if !input.HoraFin.IsZero() {
		horario.HoraFin = input.HoraFin
	}

	// Validar que la hora de fin sea mayor que la de inicio
	if horario.HoraFin.Before(horario.HoraInicio) || horario.HoraFin.Equal(horario.HoraInicio) {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "La hora de fin debe ser posterior a la hora de inicio")
		return
	}

	if err := tx.Save(&horario).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al actualizar horario: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos actualizados para la respuesta
	if err := initializers.GetDB().
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&horario, horario.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, horario)
}

// DeleteHorario elimina un horario
func DeleteHorario(c *gin.Context) {
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

	result := tx.Delete(&models.Horario{}, id)
	if result.Error != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al eliminar horario: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusNotFound, "Horario no encontrado")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "Horario eliminado correctamente"})
}