package models

import "time"

type Notificacion struct {
    ID         uint      `gorm:"primaryKey"`
    IDUsuario  uint      `gorm:"not null"`
    Usuario    Usuario   `gorm:"foreignKey:IDUsuario"` // Relación con Usuario
    CitaID     uint      `gorm:"not null"`
    Cita       Cita      `gorm:"foreignKey:CitaID"` // Relación con Cita
    Tipo       string    `gorm:"type:varchar(20);check(tipo IN ('confirmación', 'recordatorio', 'cancelación'))"`
    Mensaje    string    `gorm:"type:text"`
    FechaEnvio time.Time `gorm:"not null"`
}