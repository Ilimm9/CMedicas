package dto

type PersonaInput struct {
	Nombre          string `json:"nombre" binding:"required"`
	ApellidoPaterno string `json:"apellidoPaterno" binding:"required"`
	ApellidoMaterno string `json:"apellidoMaterno" binding:"required"`
	Telefono        string `json:"telefono"`
	FechaNacimiento string `json:"fechaNacimiento" binding:"required"`
	Genero          string `json:"genero" binding:"required"`
	Direccion       string `json:"direccion"`
}

type PersonaResponse struct {
	ID              uint   `json:"id"`
	Nombre          string `json:"nombre"`
	ApellidoPaterno string `json:"apellidoPaterno"`
	ApellidoMaterno string `json:"apellidoMaterno"`
	Telefono        string `json:"telefono"`
	FechaNacimiento string `json:"fechaNacimiento"`
	Genero          string `json:"genero"`
	Direccion       string `json:"direccion"`
}
