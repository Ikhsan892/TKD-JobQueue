package postgresql

import (
	"assessment/pkg/entity"
	"gorm.io/gorm"
)

type IProjectRepo interface {
	GetById(projectId uint) (entity.GetProject, error)
}

type ProjectPostgresAdapter struct {
	db *gorm.DB
}

func NewProjectAdapter(db *gorm.DB) *ProjectPostgresAdapter {
	return &ProjectPostgresAdapter{
		db: db,
	}
}

func (p *ProjectPostgresAdapter) GetById(projectId uint) (entity.GetProject, error) {
	var result entity.GetProject

	if err := p.db.Table("projects").Where("id = ?", projectId).First(&result).Error; err != nil {
		return entity.GetProject{}, err
	}

	return result, nil
}
