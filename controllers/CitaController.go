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

type CitaInput struct {
	PacienteID uint      `json:"paciente_id" binding:"required"`
	MedicoID   uint      `json:"medico_id" binding:"required"`
	FechaCita  time.Time `json:"fecha_cita" binding:"required"`
	Motivo     string    `json:"motivo" binding:"required,max=500"`
}

// Crear una nueva cita
func PostCita(c *gin.Context) {
	var input CitaInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verificar que el paciente existe
	var paciente models.Usuario
	if err := initializers.GetDB().First(&paciente, input.PacienteID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusBadRequest, "Paciente no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al verificar paciente: "+err.Error())
		}
		return
	}

	// Verificar que el médico existe
	var medico models.Medico
	if err := initializers.GetDB().First(&medico, input.MedicoID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusBadRequest, "Médico no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al verificar médico: "+err.Error())
		}
		return
	}

	// Validar que la fecha sea futura
	if input.FechaCita.Before(time.Now()) {
		utils.RespondError(c, http.StatusBadRequest, "La fecha de la cita debe ser futura")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	cita := models.Cita{
		PacienteID: input.PacienteID,
		MedicoID:   input.MedicoID,
		FechaCita:  input.FechaCita,
		Motivo:     input.Motivo,
		Estado:     "programada",
	}

	if err := tx.Create(&cita).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al guardar cita: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar relaciones para la respuesta
	if err := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&cita, cita.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos de la cita: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, cita)
}

// Obtener una cita por ID
func GetCita(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var cita models.Cita
	result := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		Preload("Notificaciones").
		First(&cita, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Cita no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar cita: "+result.Error.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, cita)
}

// Obtener todas las citas
func GetAllCitas(c *gin.Context) {
	var citas []models.Cita
	result := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		Find(&citas)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener citas: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, citas)
}

// GetCitasUsuarioActual obtiene las citas del usuario autenticado según su rol
func GetCitasUsuarioActual(c *gin.Context) {
	// Obtener información del usuario autenticado
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "No se pudo identificar al usuario")
		return
	}

	userRol, exists := c.Get("userRol")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "No se pudo verificar el rol del usuario")
		return
	}

	var citas []models.Cita
	query := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona")

	// Filtrar según el rol del usuario
	switch userRol {
	case "paciente":
		query = query.Where("paciente_id = ?", userID)
	case "medico":
		// Primero obtener el ID del médico asociado a este usuario
		var medico models.Medico
		if err := initializers.GetDB().Where("usuario_id = ?", userID).First(&medico).Error; err != nil {
			utils.RespondError(c, http.StatusNotFound, "No se encontró médico asociado a este usuario")
			return
		}
		query = query.Where("medico_id = ?", medico.ID)
	case "administrador":
		// Administradores ven todas las citas sin filtro
	default:
		utils.RespondError(c, http.StatusForbidden, "Rol no autorizado para ver citas")
		return
	}

	// Aplicar filtros opcionales
	if estado := c.Query("estado"); estado != "" {
		query = query.Where("estado = ?", estado)
	}

	if fecha := c.Query("fecha"); fecha != "" {
		// Formato esperado: YYYY-MM-DD
		if _, err := time.Parse("2006-01-02", fecha); err != nil {
			utils.RespondError(c, http.StatusBadRequest, "Formato de fecha inválido. Use YYYY-MM-DD")
			return
		}
		query = query.Where("DATE(fecha_cita) = ?", fecha)
	}

	// Ordenar por fecha de cita (más recientes primero)
	query = query.Order("fecha_cita DESC")

	if err := query.Find(&citas).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener citas: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, citas)
}

// Actualizar una cita existente
func UpdateCita(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input struct {
		FechaCita *time.Time `json:"fecha_cita"`
		Motivo    string     `json:"motivo" binding:"max=500"`
		Estado    string     `json:"estado" binding:"omitempty,oneof=programada cancelada completada"`
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

	var cita models.Cita
	if err := tx.First(&cita, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Cita no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar cita: "+err.Error())
		}
		return
	}

	// Actualizar solo info dada
	if input.FechaCita != nil {
		if input.FechaCita.Before(time.Now()) {
			tx.Rollback()
			utils.RespondError(c, http.StatusBadRequest, "La fecha de la cita debe ser futura")
			return
		}
		cita.FechaCita = *input.FechaCita
	}
	if input.Motivo != "" {
		cita.Motivo = input.Motivo
	}
	if input.Estado != "" {
		cita.Estado = input.Estado
	}

	if err := tx.Save(&cita).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al actualizar cita: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos actualizados para la respuesta
	if err := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&cita, cita.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, cita)
}

// Eliminar una cita
func DeleteCita(c *gin.Context) {
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

	// Verificar si la cita tiene notificaciones asociadas
	var count int64
	if err := tx.Model(&models.Notificacion{}).Where("cita_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al verificar notificaciones: "+err.Error())
		return
	}

	if count > 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "No se puede eliminar, la cita tiene notificaciones asociadas")
		return
	}

	result := tx.Delete(&models.Cita{}, id)
	if result.Error != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al eliminar cita: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusNotFound, "Cita no encontrada")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "Cita eliminada correctamente"})
}

// Obtener las citas de un paciente específico
func GetCitasPorPaciente(c *gin.Context) {
	pacienteID, err := strconv.Atoi(c.Param("paciente_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID de paciente inválido")
		return
	}

	var citas []models.Cita
	result := initializers.GetDB().
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		Where("paciente_id = ?", pacienteID).
		Find(&citas)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener citas: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, citas)
}

// Obtener las citas de un médico específico
func GetCitasPorMedico(c *gin.Context) {
	medicoID, err := strconv.Atoi(c.Param("medico_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID de médico inválido")
		return
	}

	var citas []models.Cita
	result := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Where("medico_id = ?", medicoID).
		Find(&citas)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener citas: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, citas)
}

// Cancelar una cita existente
func CancelarCita(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	// informacion del usuario 
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "No se pudo identificar al usuario")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	var cita models.Cita
	if err := tx.Preload("Paciente").First(&cita, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Cita no encontrada")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar cita: "+err.Error())
		}
		return
	}

	// Verificar permisos, el usuario o admi son los unicos que pueden cancelar
	if cita.PacienteID != userID.(uint) {
		userRol, _ := c.Get("userRol")
		if userRol != "administrador" {
			tx.Rollback()
			utils.RespondError(c, http.StatusForbidden, "No tienes permiso para cancelar esta cita")
			return
		}
	}

	// Validar que la cita no esté ya cancelada o completada
	if cita.Estado == "cancelada" {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "La cita ya está cancelada")
		return
	}

	if cita.Estado == "completada" {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "No se puede cancelar una cita ya completada")
		return
	}

	// Validar que no se cancele con muy poca anticipación (< de 24 horas)
	if time.Until(cita.FechaCita) < 24*time.Hour {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "No se puede cancelar con menos de 24 horas de anticipación")
		return
	}

	// Actualizar estado de la cita
	cita.Estado = "cancelada"
	if err := tx.Save(&cita).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al cancelar cita: "+err.Error())
		return
	}

	// Crear notificación de cancelación
	notificacion := models.Notificacion{
		IDUsuario:  cita.PacienteID,
		CitaID:     cita.ID,
		Tipo:       "cancelación",
		Mensaje:    "Su cita ha sido cancelada",
		FechaEnvio: time.Now(),
	}

	if err := tx.Create(&notificacion).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al crear notificación: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	if err := initializers.GetDB().
		Preload("Paciente").
		Preload("Paciente.Persona").
		Preload("Medico").
		Preload("Medico.Usuario").
		Preload("Medico.Usuario.Persona").
		First(&cita, cita.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Cita cancelada exitosamente",
		"cita":    cita,
	})
}