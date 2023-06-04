package entity

type GetVolumeAttachment struct {
	Id          uint   `json:"id"`
	ProjectId   uint   `json:"project_id" gorm:"not null"`
	StructureId string `json:"structure_id" gorm:"not null"`
	VolumeId    uint   `json:"volume_id" gorm:"not null"`
	FileName    string `json:"file_name" gorm:"not null"`
	Path        string `json:"path" gorm:"not null"`
	Format      string `json:"format" gorm:"not null"`
}

type GetVolume struct {
	Id                           uint    `json:"id"`
	ProjectId                    uint    `json:"project_id" gorm:"not null"`
	StructureId                  string  `json:"structure_id" gorm:"not null"`
	StorageFacilityId            uint    `json:"storage_facility_id" gorm:"not null"`
	StorageFacilityName          string  `json:"storage_facility_name"`
	FilledByName                 string  `json:"filled_by_name"`
	Unit                         float64 `json:"unit" gorm:"not null"`
	Shelf                        float64 `json:"shelf" gorm:"not null"`
	ShelfLong                    float64 `json:"shelf_long" gorm:"not null"`
	VolumeStorageFacility        float64 `json:"volume_storage_facility" gorm:"not null"`
	VolumeStorageMediaPercentage float64 `json:"volume_storage_media_percentage" gorm:"not null"`
	VolumeStorageMedia           float32 `json:"volume_storage_media" gorm:"not null"`
	FilledBy                     uint    `json:"filled_by" gorm:"not null"`
}

type VolumeStats struct {
	Type                string  `json:"type"`
	TotalFacilityVolume float64 `json:"total_facility_volume"`
	TotalMediaVolume    float64 `json:"total_media_volume"`
}
