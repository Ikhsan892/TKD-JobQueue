package postgresql

import "gorm.io/gorm"

type InventoryPostgreAdapter struct {
	db *gorm.DB
}

func NewInventoryAdapter(db *gorm.DB) *InventoryPostgreAdapter {
	return &InventoryPostgreAdapter{
		db: db,
	}
}

func (i *InventoryPostgreAdapter) GetMotherfucker() {

}
