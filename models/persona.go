package models

import "time"

type Persona struct {
    ID              uint      `gorm:"primaryKey"`
    Nombre          string    `gorm:"size:100;not null"`
    ApellidoPaterno string    `gorm:"size:100;not null"`
    ApellidoMaterno string    `gorm:"size:100;not null"`
    Telefono        string    `gorm:"size:15"`
    FechaNacimiento time.Time
    Genero          string    `gorm:"type:varchar(20);check(genero IN ('masculino', 'femenino', 'otro'))"`
    Direccion       string    `gorm:"type:text"`
    
}
