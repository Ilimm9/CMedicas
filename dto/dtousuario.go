package dto

type UsuarioInput struct {
	PersonaID  uint   `json:"persona_id" binding:"required"`
	Rol        string `json:"rol" binding:"required,oneof=paciente medico administrador"`
	Correo     string `json:"correo" binding:"required,email"`
	Contrasena string `json:"contrasena" binding:"required,min=8"`
}

type UsuarioResponse struct {
	ID        uint   `json:"id"`
	Correo    string `json:"correo"`
	Rol       string `json:"rol"`
	PersonaID uint   `json:"persona_id"`
}
