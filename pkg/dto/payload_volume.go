package dto

type PayloadVolume struct {
	UserId    uint    `json:"user_id"`
	CompanyId uint    `json:"company_id"`
	ProjectId uint    `json:"project_id"`
	UnitKerja *string `json:"unit_kerja"`
}

type PayloadInventory struct {
	PayloadVolume
}
