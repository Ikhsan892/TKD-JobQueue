package dto

type ReportVolume struct {
	ProjectId   uint   `json:"project_id"`
	StructureId string `json:"structure_id"`
	UserId      uint   `json:"user_id"`
}
