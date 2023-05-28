package postgresql

import (
	"assessment/pkg/entity"
	"gorm.io/gorm"
)

type IVolumeRepo interface {
	GetAttachmentsByVolumeId(volumeId uint) []entity.GetVolumeAttachment
	GetById(volumeId uint) (entity.GetVolume, error)
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
