package dto

type GetQuestion struct {
	ProjectId        uint
	CompanyId        uint
	UnitKerja        *string
	StatusPertanyaan *uint
	ProjectType      *uint
}
