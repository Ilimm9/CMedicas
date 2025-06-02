package models

type Medico struct {
    ID           uint    `gorm:"primaryKey"`
    UsuarioID    uint    `gorm:"unique;not null"`
    Usuario      Usuario `gorm:"foreignKey:UsuarioID"`
    Especialidad string  `gorm:"size:100;not null"`
    Horarios    []Horario `gorm:"foreignKey:MedicoID"`
    Cita       []Cita    `gorm:"foreignKey:MedicoID"` 
}