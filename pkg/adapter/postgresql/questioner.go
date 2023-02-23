package postgresql

import (
	"assessment/pkg/entity"
	"gorm.io/gorm"
)

type QuestionerPostgreAdapter struct {
	db *gorm.DB
}

func NewQuestionerAdapter(db *gorm.DB) *QuestionerPostgreAdapter {
	return &QuestionerPostgreAdapter{
		db: db,
	}
}

func (questioner *QuestionerPostgreAdapter) GetAllQuestion(projectId, templateId uint) []entity.GetQuestion {
	var result []entity.GetQuestion

	questioner.db.Table("project_questioners").
		Where("project_id", projectId).
		Where("template_id", templateId).
		Order("id asc").
		Select("distinct id,name").Find(&result)

	return result
}

func (questioner *QuestionerPostgreAdapter) GetAllUnitKerja(projectId uint) []entity.GetUnitKerja {
	var result []entity.GetUnitKerja

	questioner.db.Table("project_company_structures").
		Where("project_id = ?", projectId).
		Select("uuid as id,name").
		Order("name asc").
		Find(&result)

	return result
}

func (questioner *QuestionerPostgreAdapter) GetAnswers(projectId, templateId uint, unitId string) []string {
	var result []struct {
		Answer string `json:"answer"`
	}

	questioner.db.Table("project_questioners").
		Joins("left join project_questioner_answereds answered on answered.project_questioner_id = project_questioners.id").
		Where("project_questioners.project_id = ?", projectId).
		Where("project_questioners.template_id = ?", templateId).
		Where("project_questioners.project_company_structure_id = ?", unitId).
		Select(`
			concat(
				(case 
					when answered.answer_value = 1 then 'Ya'
					when answered.answer_value = 2 then ''
					when answered.answer_value = 0 then 'Tidak'
				 end
				),
				(
					SELECT ' ; ' || string_agg( concat(arr.item_object->>'label',' : ',arr.item_object->>'value')::TEXT,' , ')
					FROM jsonb_array_elements(
						(
							case jsonb_typeof(answered.conditions)
								when 'array' then answered.conditions 
								else '[{"label" : "","value" : ""}]' 
							end
						)
					) with ordinality arr(item_object, position)
				),
				(case 
					when answered.answer_value = 2 then ' ; description : ' || answered.description
				 end
				),
				' ; keterangan : ' || answered.note
			) as answer
		`).
		Order("project_questioners.id asc").
		Find(&result)

	var s []string
	for _, r := range result {
		s = append(s, r.Answer)
	}

	return s
}

func (questioner *QuestionerPostgreAdapter) GetTemplateQuestionOnProject(projectId uint) []entity.GetTemplate {
	var (
		result      []entity.GetTemplate
		templateIds []uint
	)

	questioner.db.Table("project_questioners").
		Where("project_id = ?", projectId).
		Distinct("template_id").
		Select("template_id").
		Find(&templateIds)

	questioner.db.Table("question_templates").
		Where("id in ?", templateIds).
		Select("id,name").
		Find(&result)

	return result
}

func (questioner *QuestionerPostgreAdapter) GetProjectDetail(projectId uint) (entity.GetProjectDetail, error) {
	var project entity.GetProjectDetail

	if err := questioner.db.Table("projects").
		Where("id = ?", projectId).
		Select("id,name,code").
		First(&project).Error; err != nil {
		return entity.GetProjectDetail{}, err
	}

	return project, nil
}
