package postgresql

import (
	"assessment/pkg/entity"
	"gorm.io/gorm"
)

type IVolumeRepo interface {
	GetAttachmentsByVolumeId(volumeId uint) []entity.GetVolumeAttachment
	GetById(volumeId uint) (entity.GetVolume, error)
	GetByStructureId(structureId string) []entity.GetVolume
	GetVolumeStats(structureId string) []entity.VolumeStats
}

type VolumePostgresAdapter struct {
	db *gorm.DB
}

func NewVolumeAdapter(db *gorm.DB) *VolumePostgresAdapter {
	return &VolumePostgresAdapter{
		db: db,
	}
}

func (volume *VolumePostgresAdapter) GetById(volumeId uint) (entity.GetVolume, error) {
	var result entity.GetVolume

	if err := volume.db.Table("volumes").Where("id = ?", volumeId).First(&result).Error; err != nil {
		return entity.GetVolume{}, err
	}

	return result, nil
}

func (volume *VolumePostgresAdapter) GetAttachmentsByVolumeId(volumeId uint) []entity.GetVolumeAttachment {
	var result []entity.GetVolumeAttachment

	volume.db.Table("volume_attachments").Where("volume_id = ?", volumeId).Find(&result)

	return result
}

func (volume *VolumePostgresAdapter) GetByStructureId(structureId string) []entity.GetVolume {
	var result []entity.GetVolume

	volume.db.Table("volumes").
		Where("structure_id = ?", structureId).
		Joins("inner join users on users.id = volumes.filled_by").
		Joins("inner join storage_facilities on storage_facilities.id = volumes.storage_facility_id").
		Select("volumes.*,users.full_name as filled_by_name,storage_facilities.name as storage_facility_name").
		Find(&result)

	return result
}

func (volume *VolumePostgresAdapter) GetVolumeStats(structureId string) []entity.VolumeStats {
	var result []entity.VolumeStats

	volume.db.Table("volumes").
		Where("structure_id = ?", structureId).
		Joins("inner join storage_facilities on storage_facilities.id = volumes.storage_facility_id").
		Select(`
				storage_facilities.name as type, 
				sum(volumes.volume_storage_facility) as total_facility_volume,
				sum(volumes.volume_storage_media) as total_media_volume 
		`).
		Group("storage_facilities.name").
		Find(&result)

	return result
}
