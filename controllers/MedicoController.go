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

type MedicoInput struct {
	UsuarioID    uint   `json:"usuario_id" binding:"required"`
	Especialidad string `json:"especialidad" binding:"required,max=100"`
}

// PostMedico crea un nuevo médico
func PostMedico(c *gin.Context) {
	var input MedicoInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verificar si el usuario existe y es de tipo médico
	var usuario models.Usuario
	if err := initializers.GetDB().First(&usuario, input.UsuarioID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusBadRequest, "Usuario no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al verificar usuario: "+err.Error())
		}
		return
	}

	if usuario.Rol != "medico" {
		utils.RespondError(c, http.StatusBadRequest, "El usuario debe tener rol 'medico'")
		return
	}

	tx := initializers.GetDB().Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al iniciar transacción: "+tx.Error.Error())
		return
	}

	medico := models.Medico{
		UsuarioID:    input.UsuarioID,
		Especialidad: input.Especialidad,
	}

	if err := tx.Create(&medico).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al guardar médico: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos relacionados para la respuesta
	if err := initializers.GetDB().Preload("Usuario").Preload("Usuario.Persona").First(&medico, medico.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos del médico: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, medico)
}

// GetMedico obtiene un médico por ID
func GetMedico(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var medico models.Medico
	result := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Preload("Horarios").
		First(&medico, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Médico no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar médico: "+result.Error.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, medico)
}

// GetAllMedicos obtiene todos los médicos
func GetAllMedicos(c *gin.Context) {
	var medicos []models.Medico
	result := initializers.GetDB().
		Preload("Usuario").
		Preload("Usuario.Persona").
		Find(&medicos)

	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener médicos: "+result.Error.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, medicos)
}

// UpdateMedico actualiza un médico existente
func UpdateMedico(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input struct {
		Especialidad string `json:"especialidad" binding:"max=100"`
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

	var medico models.Medico
	if err := tx.First(&medico, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "Médico no encontrado")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, "Error al buscar médico: "+err.Error())
		}
		return
	}

	// Actualizar solo los campos proporcionados
	if input.Especialidad != "" {
		medico.Especialidad = input.Especialidad
	}

	if err := tx.Save(&medico).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al actualizar médico: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	// Cargar datos actualizados para la respuesta
	if err := initializers.GetDB().Preload("Usuario").Preload("Usuario.Persona").First(&medico, medico.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al cargar datos actualizados: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, medico)
}

// DeleteMedico elimina un médico
func DeleteMedico(c *gin.Context) {
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

	// Verificar si el médico tiene horarios o citas asociadas
	var count int64
	if err := tx.Model(&models.Horario{}).Where("medico_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al verificar horarios: "+err.Error())
		return
	}

	if count > 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "No se puede eliminar, el médico tiene horarios asignados")
		return
	}

	if err := tx.Model(&models.Cita{}).Where("medico_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al verificar citas: "+err.Error())
		return
	}

	if count > 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "No se puede eliminar, el médico tiene citas programadas")
		return
	}

	result := tx.Delete(&models.Medico{}, id)
	if result.Error != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Error al eliminar médico: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusNotFound, "Médico no encontrado")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al confirmar transacción: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "Médico eliminado correctamente"})
}

// Obtener lista de médicos disponibles, info basica
func GetMedicosDisponibles(c *gin.Context) {
	// Obtener parámetros de consulta opcionales
	especialidad := c.Query("especialidad")
	fecha := c.Query("fecha") // Formato esperado: YYYY-MM-DD

	var medicos []struct {
		ID           uint   `json:"id"`
		Nombre       string `json:"nombre"`
		Apellidos    string `json:"apellidos"`
		Especialidad string `json:"especialidad"`
		FotoPerfil   string `json:"foto_perfil,omitempty"`
	}

	query := initializers.GetDB().
		Model(&models.Medico{}).
		Select("medicos.id, personas.nombre, personas.apellido_paterno || ' ' || personas.apellido_materno as apellidos, medicos.especialidad, usuarios.foto_perfil").
		Joins("JOIN usuarios ON usuarios.id = medicos.usuario_id").
		Joins("JOIN personas ON personas.id = usuarios.persona_id").
		Where("medicos.activo = ?", true)

	// Filtro por especialidad si se proporciona
	if especialidad != "" {
		query = query.Where("medicos.especialidad ILIKE ?", "%"+especialidad+"%")
	}

	// Filtro por disponibilidad en fecha específica si se proporciona
	if fecha != "" {
		parsedDate, err := time.Parse("2006-01-02", fecha)
		if err != nil {
			utils.RespondError(c, http.StatusBadRequest, "Formato de fecha inválido. Use YYYY-MM-DD")
			return
		}

		diaSemana := parsedDate.Weekday().String()
		if diaSemana == "Sunday" {
			diaSemana = "Domingo"
		} // Adaptar según tus necesidades

		query = query.
			Joins("JOIN horarios ON horarios.medico_id = medicos.id").
			Where("horarios.dia_semana = ?", diaSemana)
	}

	if err := query.Find(&medicos).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Error al obtener médicos disponibles: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, medicos)
}