package postgresql

import (
	"assessment/pkg/entity"
	"gorm.io/gorm"
)

type ICompanyStructureRepo interface {
	GetById(structureId string) (entity.CompanyStructure, error)
	GetByProject(projectId uint) []entity.CompanyStructure
}

type CompanyStructurePostgresAdapter struct {
	db *gorm.DB
}

func NewCompanyStructureAdapter(db *gorm.DB) *CompanyStructurePostgresAdapter {
	return &CompanyStructurePostgresAdapter{
		db: db,
	}
}

func (p *CompanyStructurePostgresAdapter) GetByProject(projectId uint) []entity.CompanyStructure {
	var result []entity.CompanyStructure

	if err := p.db.Table("project_company_structures").Where("project_id = ?", projectId).Find(&result).Error; err != nil {
		return []entity.CompanyStructure{}
	}

	return result
}

func (p *CompanyStructurePostgresAdapter) GetById(structureId string) (entity.CompanyStructure, error) {
	var result entity.CompanyStructure

	if err := p.db.Table("project_company_structures").Where("uuid = ?", structureId).First(&result).Error; err != nil {
		return entity.CompanyStructure{}, err
	}

	return result, nil
}
