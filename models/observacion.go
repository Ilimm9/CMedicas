package models

import "time"

type Observacion struct {
    ID            uint      `gorm:"primaryKey"`
    CitaID        uint      `gorm:"not null;uniqueIndex"` // Una observación por cita
    Cita          Cita      `gorm:"foreignKey:CitaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    Observaciones string    `gorm:"type:text"`
    Diagnostico   string    `gorm:"type:text"`
    FechaRegistro time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}