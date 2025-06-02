package models

import "time"

type Cita struct {
    ID         uint      `gorm:"primaryKey"`
    PacienteID uint      `gorm:"not null"`
    Paciente   Usuario   `gorm:"foreignKey:PacienteID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    MedicoID   uint      `gorm:"not null"`
    Medico     Medico    `gorm:"foreignKey:MedicoID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    FechaCita  time.Time `gorm:"not null;index"` // Índice para búsquedas
    Motivo     string    `gorm:"type:text"`
    Estado     string    `gorm:"type:varchar(20);check(estado IN ('programada', 'cancelada', 'completada'));index"`
    CreadaEn   time.Time `gorm:"autoCreateTime"`
    
    Notificaciones []Notificacion `gorm:"foreignKey:CitaID"`
}
