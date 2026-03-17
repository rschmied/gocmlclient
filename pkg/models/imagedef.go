// Package models provides the models for Cisco Modeling Labs
// here: image definition related types
package models

// ImageDefinition represents an image definition known to the controller.
type ImageDefinition struct {
	ID            UUID   `json:"id"`
	SchemaVersion string `json:"schema_version,omitempty"`
	NodeDefID     string `json:"node_definition_id"`
	Description   string `json:"description,omitempty"`
	Label         string `json:"label"`

	DiskImage  *string `json:"disk_image,omitempty"`
	DiskImage2 *string `json:"disk_image_2,omitempty"`
	DiskImage3 *string `json:"disk_image_3,omitempty"`
	DiskImage4 *string `json:"disk_image_4,omitempty"`

	ReadOnly      bool              `json:"read_only"`
	Configuration *string           `json:"configuration,omitempty"`
	DockerTag     *string           `json:"docker_tag,omitempty"`
	EFIBoot       bool              `json:"efi_boot,omitempty"`
	PyATS         *PyAtsCredentials `json:"pyats,omitempty"`
	SHA256        *string           `json:"sha256,omitempty"`

	RAM          *int `json:"ram,omitempty"`
	CPUs         *int `json:"cpus,omitempty"`
	CPUlimit     *int `json:"cpu_limit,omitempty"`
	DataVolume   *int `json:"data_volume,omitempty"`
	BootDiskSize *int `json:"boot_disk_size,omitempty"`
}

// DiskImage1 is a legacy accessor for the primary disk image.
// Prefer using DiskImage going forward.
func (id ImageDefinition) DiskImage1() string {
	if id.DiskImage == nil {
		return ""
	}
	return *id.DiskImage
}
