package dto

type PayloadQuestioner struct {
	UserId           uint    `json:"user_id"`
	CompanyId        uint    `json:"company_id"`
	ProjectId        uint    `json:"project_id"`
	UnitKerja        *string `json:"unit_kerja"`
	StatusPertanyaan *uint   `json:"status_pertanyaan"`
	ProjectType      *uint   `json:"project_type"`
}
