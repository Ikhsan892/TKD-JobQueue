package entity

type CompanyStructure struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid" gorm:"not null"`
}
