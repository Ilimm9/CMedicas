package models 

import "time"

type Usuario struct {
    ID         uint      `gorm:"primaryKey"`
    PersonaID  uint      `gorm:"uniqueIndex;not null"`
    Persona    Persona   `gorm:"foreignKey:PersonaID"` // Referencia 
    Rol        string    `gorm:"type:varchar(20);not null;check(rol IN ('paciente','medico','administrador'))"`
    Correo     string    `gorm:"size:100;unique;not null"`
    Contrasena string    `gorm:"size:255;not null"`
    CreadoEn   time.Time `gorm:"autoCreateTime"`
    Medico      *Medico       `gorm:"foreignKey:UsuarioID"`
    Cita       []Cita        `gorm:"foreignKey:PacienteID"`
    Notificaciones []Notificacion `gorm:"foreignKey:IDUsuario"`
}