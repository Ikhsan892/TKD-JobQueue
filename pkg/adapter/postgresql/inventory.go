package postgresql

import (
	"assessment/pkg/dto"
	"gorm.io/gorm"
)

type InventoryPostgreAdapter struct {
	db *gorm.DB
}

func NewInventoryAdapter(db *gorm.DB) *InventoryPostgreAdapter {
	return &InventoryPostgreAdapter{
		db: db,
	}
}

func (i *InventoryPostgreAdapter) inventoryModel() *gorm.DB {
	return i.db.Table("inventories")
}

func (i *InventoryPostgreAdapter) GetAllData(structureId string, projectId uint) []dto.ReportInventoryResponse {
	var inventoryModel []dto.ReportInventoryResponse
	data := i.inventoryModel().Where("inventories.project_id = ?", projectId)

	data.Where("inventories.structure_id = ?", structureId)

	data.Joins("left join frequency_of_change_additions on frequency_of_change_additions.id = inventories.id").
		Joins("left join storage_media on storage_media.id = inventories.storage_media_id").
		Joins("left join storage_facilities on storage_facilities.id = inventories.storage_facilities_id").
		Joins("inner join inventory_documents on inventories.id = inventory_documents.inventory_id").
		Joins("left join document_shapes on document_shapes.id = inventory_documents.document_shape_id").
		Joins("left join dimension_size_archievs on dimension_size_archievs.id = inventory_documents.dimension_size_archiev_id").
		Joins("left join authenticity_levels on authenticity_levels.id = inventory_documents.authenticity_level_id").
		Joins("left join users on users.id = inventory_documents.filled_by").
		Order("inventories.id asc").
		Select(`
			file_number as kode_klasifikasi,
			archive_title as judul_arsip,
			frequency_of_change_additions.name as frekuensi_penambahan,
			archive_year_of as tahun_dari,
			archive_year_to as tahun_sampai,
			storage_media.name as media_simpan,
			storage_facilities.name as sarana_simpan,
			inventory_documents.document_type_in_order_of_process as isi_jenis_dokumen,
			document_shapes.name as bentuk_dokumen,
			dimension_size_archievs.name as ukuran_fisik_dimensi_arsip,
			authenticity_levels.name as tingkat_keaslian,
			users.full_name as diisi_oleh,
			users.profile_pict as profile
		`).
		Find(&inventoryModel)
	if err := data.Error; err != nil {
		return []dto.ReportInventoryResponse{}
	}

	return inventoryModel
}
