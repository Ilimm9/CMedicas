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

type NotificacionInput struct {
	IDUsuario uint   `json:"usuario_id" binding:"required"`
	CitaID    uint   `json:"cita_id" binding:"required"`
	Tipo      string `json:"tipo" binding:"required,oneof=confirmación recordatorio cancelación"`
	Mensaje   string `json:"mensaje" binding:"required,max=500"`
}

// Crear notificación
func PostNotificacion(c *gin.Context) {
	var input NotificacionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verificar que el usuario existe
	var usuario models.Usuario
	if err := initializers.GetDB().First(&usuario, input.IDUsuario).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusBadRequest, "Usuario no encontrado")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al verificar usuario: "+err.Error())
		}
		return
	}

	// Verificar que la cita existe
	var cita models.Cita
	if err := initializers.GetDB().First(&cita, input.CitaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusBadRequest, "Cita no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al verificar cita: "+err.Error())
		}
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	notificacion := models.Notificacion{
		IDUsuario:  input.IDUsuario,
		CitaID:     input.CitaID,
		Tipo:       input.Tipo,
		Mensaje:    input.Mensaje,
		FechaEnvio: time.Now(),
	}

	if err := tx.Create(&notificacion).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al guardar notificación: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	if err := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Medico").
		First(&notificacion, notificacion.ID).Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al cargar datos de la notificación: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusCreated, notificacion)
}

// Obtiener notificación por ID
func GetNotificacion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var notificacion models.Notificacion
	result := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Medico").
		First(&notificacion, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusNotFound, "Notificación no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar notificación: "+result.Error.Error())
		}
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, notificacion)
}

// Obtiener las notificaciones de un usuario
func GetNotificacionesPorUsuario(c *gin.Context) {
	usuarioID, err := strconv.Atoi(c.Param("usuario_id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	var notificaciones []models.Notificacion
	result := initializers.GetDB().
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Medico").
		Where("id_usuario = ?", usuarioID).
		Order("fecha_envio DESC").
		Find(&notificaciones)

	if result.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al obtener notificaciones: "+result.Error.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, notificaciones)
}

// Obtener notificaciones de una cita específica
func GetNotificacionesPorCita(c *gin.Context) {
	citaID, err := strconv.Atoi(c.Param("cita_id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID de cita inválido")
		return
	}

	var notificaciones []models.Notificacion
	result := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Where("cita_id = ?", citaID).
		Order("fecha_envio DESC").
		Find(&notificaciones)

	if result.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al obtener notificaciones: "+result.Error.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, notificaciones)
}

// Actualizar notificación
func UpdateNotificacion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input struct {
		Tipo    string `json:"tipo" binding:"omitempty,oneof=confirmación recordatorio cancelación"`
		Mensaje string `json:"mensaje" binding:"omitempty,max=500"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	var notificacion models.Notificacion
	if err := tx.First(&notificacion, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusNotFound, "Notificación no encontrada")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar notificación: "+err.Error())
		}
		return
	}

	if input.Tipo != "" {
		notificacion.Tipo = input.Tipo
	}
	if input.Mensaje != "" {
		notificacion.Mensaje = input.Mensaje
	}

	if err := tx.Save(&notificacion).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al actualizar notificación: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	if err := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Preload("Cita").
		Preload("Cita.Paciente").
		Preload("Cita.Medico").
		First(&notificacion, notificacion.ID).Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, notificacion)
}

// Eliminar notificación
func DeleteNotificacion(c *gin.Context) {
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

	result := tx.Delete(&models.Notificacion{}, id)
	if result.Error != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al eliminar notificación: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusNotFound, "Notificación no encontrada")
		return
	}

	if err := tx.Commit().Error; err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, gin.H{"message": "Notificación eliminada correctamente"})
}
