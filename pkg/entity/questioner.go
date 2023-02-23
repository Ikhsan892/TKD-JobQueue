package entity

type GetQuestion struct {
	Id   uint
	Name string
}

type GetUnitKerja struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetTemplate struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	ProjectName string
}

type GetAnswers struct {
	Id        uint   `json:"id"`
	UnitKerja string `json:"unit_kerja"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
}

type GetProjectDetail struct {
	Id   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
