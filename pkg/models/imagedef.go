// Package models provides the models for Cisco Modeling Labs
// here: image definition related types
package models

// ImageDefinition represents an image definition known to the controller.
type ImageDefinition struct {
	ID            UUID    `json:"id"`
	SchemaVersion string  `json:"schema_version"`
	NodeDefID     string  `json:"node_definition_id"`
	Description   string  `json:"description"`
	Label         string  `json:"label"`
	DiskImage1    string  `json:"disk_image"`
	DiskImage2    *string `json:"disk_image_2,omitempty"`
	DiskImage3    *string `json:"disk_image_3,omitempty"`
	ReadOnly      bool    `json:"read_only"`
	DiskSubfolder string  `json:"disk_subfolder"`
	RAM           *int    `json:"ram,omitempty"`
	CPUs          *int    `json:"cpus,omitempty"`
	CPUlimit      *int    `json:"cpu_limit,omitempty"`
	DataVolume    *int    `json:"data_volume,omitempty"`
	BootDiskSize  *int    `json:"boot_disk_size,omitempty"`
}
