package models

import "time"

type Horario struct {
    ID         uint   `gorm:"primaryKey"`
    MedicoID   uint   `gorm:"not null"`
    Medico     Medico `gorm:"foreignKey:MedicoID"` // Relación con Médico
    DiaSemana  string `gorm:"type:varchar(15);check(dia_semana IN ('Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo'))"`
    HoraInicio time.Time
    HoraFin    time.Time
}
