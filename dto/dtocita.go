package dto
 
import "time"

type CitaInput struct {
	PacienteID uint      `json:"paciente_id" binding:"required"`
	MedicoID   uint      `json:"medico_id" binding:"required"`
	FechaCita  time.Time `json:"fecha_cita" binding:"required"`
	Motivo     string    `json:"motivo" binding:"required,max=500"`
}
