package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Ilimm9/CMedicas/Repositories"
	"github.com/Ilimm9/CMedicas/Respuestas"
	"github.com/Ilimm9/CMedicas/clave"
	"github.com/Ilimm9/CMedicas/dto"
	"github.com/Ilimm9/CMedicas/initializers"
	"github.com/Ilimm9/CMedicas/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PostUsuario(c *gin.Context) {
	var input dto.UsuarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if !repositories.ExistePersonaPorID(input.PersonaID) {
		respuestas.RespondError(c, http.StatusBadRequest, "Persona no encontrada")
		return
	}

	if repositories.ExisteCorreo(input.Correo) {
		respuestas.RespondError(c, http.StatusBadRequest, "Correo ya registrado")
		return
	}

	hashedPassword, err := clave.HashPassword(input.Contrasena)
	if err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al hashear contraseña")
		return
	}

	usuario := models.Usuario{
		PersonaID:  input.PersonaID,
		Rol:        input.Rol,
		Correo:     input.Correo,
		Contrasena: hashedPassword,
	}

	if err := repositories.CrearUsuario(&usuario); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response := dto.UsuarioResponse{
		ID:        usuario.ID,
		PersonaID: usuario.PersonaID,
		Rol:       usuario.Rol,
		Correo:    usuario.Correo,
	}

	respuestas.RespondSuccess(c, http.StatusCreated, response)
}

func RegistroCompleto(c *gin.Context) {
	// 1. Estructura para el input
	var input struct {
		Nombre          string `json:"nombre" binding:"required"`
		ApellidoPaterno string `json:"apellido_paterno" binding:"required"`
		ApellidoMaterno string `json:"apellido_materno" binding:"required"`
		Correo          string `json:"correo" binding:"required,email"`
		Telefono        string `json:"telefono"`
		FechaNacimiento string `json:"fecha_nacimiento" binding:"required"`
		Genero          string `json:"genero" binding:"required,oneof=masculino femenino otro"`
		Direccion       string `json:"direccion"`
		Contrasena      string `json:"contrasena" binding:"required,min=8"`
	}

	// 2. Validar el input
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Parsear fecha (formato: "dd/mm/aaaa")
	fechaNac, err := time.Parse("02/01/2006", input.FechaNacimiento)
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "Formato de fecha inválido. Use dd/mm/aaaa")
		return
	}

	// 4. Iniciar transacción de BD
	tx := initializers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 5. Crear Persona
	persona := models.Persona{
		Nombre:          input.Nombre,
		ApellidoPaterno: input.ApellidoPaterno,
		ApellidoMaterno: input.ApellidoMaterno,
		Telefono:        input.Telefono,
		FechaNacimiento: fechaNac,
		Genero:          input.Genero,
		Direccion:       input.Direccion,
	}

	if err := tx.Create(&persona).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al crear persona")
		return
	}

	// 6. Crear Usuario (rol "paciente" por defecto)
	hashedPassword, _ := clave.HashPassword(input.Contrasena)
	usuario := models.Usuario{
		PersonaID:  persona.ID,
		Correo:     input.Correo,
		Contrasena: hashedPassword,
		Rol:        "paciente", // Rol por defecto
	}

	if err := tx.Create(&usuario).Error; err != nil {
		tx.Rollback()
		respuestas.RespondError(c, http.StatusConflict, "El correo ya está registrado")
		return
	}

	// 7. Commit y respuesta exitosa
	tx.Commit()

	// No devolver datos sensibles
	usuario.Contrasena = ""
	respuestas.RespondSuccess(c, http.StatusCreated, gin.H{
		"mensaje": "Registro exitoso",
		"usuario": gin.H{
			"id":     usuario.ID,
			"correo": usuario.Correo,
			"rol":    usuario.Rol,
			"persona": gin.H{
				"nombre_completo": persona.Nombre + " " + persona.ApellidoPaterno,
			},
		},
	})
}

// Obtener usuario por ID
func GetUsuario(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	usuario, err := repositories.ObtenerUsuarioPorID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuestas.RespondError(c, http.StatusNotFound, "Usuario no encontrado")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := dto.UsuarioResponse{
		ID:        usuario.ID,
		PersonaID: usuario.PersonaID,
		Rol:       usuario.Rol,
		Correo:    usuario.Correo,
	}

	respuestas.RespondSuccess(c, http.StatusOK, response)
}



// Obtener todos los usuarios
func GetAllUsuarios(c *gin.Context) {
	var usuarios []models.Usuario
	result := initializers.GetDB().Preload("Persona").Find(&usuarios)
	if result.Error != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al obtener usuarios: "+result.Error.Error())
		return
	}

	// Ocultar contraseñas
	for i := range usuarios {
		usuarios[i].Contrasena = ""
	}

	respuestas.RespondSuccess(c, http.StatusOK, usuarios)
}

// Actualizar usuario
func UpdateUsuario(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var input dto.UsuarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	usuario := models.Usuario{
		Correo:    input.Correo,
		Contrasena: input.Contrasena,
		Rol:       input.Rol,
		PersonaID: input.PersonaID,
	}

	if err := repositories.ActualizarUsuario(uint(id), &usuario); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, "Usuario actualizado correctamente")
}


// Eliminar usuario
func DeleteUsuario(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := repositories.EliminarUsuario(uint(id)); err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respuestas.RespondSuccess(c, http.StatusOK, "Usuario eliminado correctamente")
}


// Autenticar un usuario y devolver token JWT
func Login(c *gin.Context) {
	var input struct {
		Correo     string `json:"correo" binding:"required,email"`
		Contrasena string `json:"contrasena" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respuestas.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var usuario models.Usuario
	if err := initializers.GetDB().Preload("Persona").Where("correo = ?", input.Correo).First(&usuario).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuestas.RespondError(c, http.StatusUnauthorized, "Credenciales inválidas")
		} else {
			respuestas.RespondError(c, http.StatusInternalServerError, "Error al buscar usuario: "+err.Error())
		}
		return
	}

	if !clave.CheckPasswordHash(input.Contrasena, usuario.Contrasena) {
		respuestas.RespondError(c, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Generar JWT
	token, err := clave.GenerateJWT(usuario.ID, usuario.Rol)
	if err != nil {
		respuestas.RespondError(c, http.StatusInternalServerError, "Error al generar token")
		return
	}

	usuario.Contrasena = ""

	respuestas.RespondSuccess(c, http.StatusOK, gin.H{
		"token":   token,
		"usuario": usuario,
	})
}

// Obtiene información del usuario autenticado
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respuestas.RespondError(c, http.StatusUnauthorized, "No se pudo identificar al usuario")
		return
	}

	var usuario models.Usuario
	result := initializers.GetDB().Preload("Persona").First(&usuario, userID)
	if result.Error != nil {
		respuestas.RespondError(c, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	// No devolver contraseña
	usuario.Contrasena = ""
	respuestas.RespondSuccess(c, http.StatusOK, usuario)
}
