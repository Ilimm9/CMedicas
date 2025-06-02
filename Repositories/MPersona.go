package repositories

import (
	"CMedicas/models"
	"gorm.io/gorm"
)

type PersonaRepository struct {
	db *gorm.DB
}

func NuevaPersona(db *gorm.DB) *PersonaRepository {
	return &PersonaRepository{db: db}
}

func (r *PersonaRepository) Create(persona *models.Persona) error {
	return r.db.Create(persona).Error
}

func (r *PersonaRepository) ConsultarPersonas() ([]models.Persona, error) {
	var personas []models.Persona
	err := r.db.Find(&personas).Error
	return personas, err
}

func (r *PersonaRepository) ConsultarPersonaID(id uint) (models.Persona, error) {
	var persona models.Persona
	err := r.db.First(&persona, id).Error
	return persona, err
}

func (r *PersonaRepository) Update(id uint, persona models.Persona) error {
	return r.db.Model(&models.Persona{}).Where("id = ?", id).Updates(persona).Error
}

func (r *PersonaRepository) Delete(id uint) error {
	return r.db.Delete(&models.Persona{}, id).Error
}