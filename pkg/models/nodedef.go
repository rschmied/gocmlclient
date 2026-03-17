// Package models provides the models for Cisco Modeling Labs
// here: node definition related types
package models

// SimplifiedNodeDefinitionResponse represents a simplified node definition
type SimplifiedNodeDefinitionResponse struct {
	ID               string   `json:"id"`
	General          any      `json:"general"` // Keep as any for now
	Device           any      `json:"device"`  // Keep as any for now
	UI               any      `json:"ui"`      // Keep as any for now
	Sim              any      `json:"sim"`     // Keep as any for now
	ImageDefinitions []string `json:"image_definitions"`
}

// NodeDefinition represents a node definition available on the CML controller.
// The node def data structure now matches the OpenAPI schema.
type NodeDefinition struct {
	ID               UUID                        `json:"id"`
	Configuration    NodeDefinitionConfiguration `json:"configuration"`
	Device           deviceData                  `json:"device"`
	Inherited        NodeDefinitionInherited     `json:"inherited"`
	SchemaVersion    string                      `json:"schema_version"`
	Sim              simData                     `json:"sim"`
	Boot             NodeDefinitionBoot          `json:"boot"`
	PyATS            NodeDefinitionPyats         `json:"pyats"`
	General          NodeDefinitionGeneral       `json:"general"`
	UI               NodeDefinitionUI            `json:"ui"`
	ImageDefinitions []string                    `json:"image_definitions,omitempty"`
}

type simData struct {
	LinuxNative      LinuxNativeSimulation `json:"linux_native"`
	Parameters       NodeParameters        `json:"parameters"`
	UsageEstimations UsageEstimations      `json:"usage_estimations"`
	RAM              int                   `json:"ram"`
	CPUs             int                   `json:"cpus"`
	CPULimit         int                   `json:"cpu_limit"`
	DataVolume       int                   `json:"data_volume"`
	BootDiskSize     int                   `json:"boot_disk_size"`
	Console          bool                  `json:"console"`
	Simulate         bool                  `json:"simulate"`
	CustomMAC        bool                  `json:"custom_mac"`
	VNC              bool                  `json:"vnc"`
}

type deviceData struct {
	Interfaces interfaceData `json:"interfaces"`
}

type interfaceData struct {
	SerialPorts        int      `json:"serial_ports"`
	DefaultConsole     int      `json:"default_console"`
	Physical           []string `json:"physical"`
	HasLoopbackZero    bool     `json:"has_loopback_zero"`
	MinCount           int      `json:"min_count"`
	DefaultCount       int      `json:"default_count"`
	IOLStaticEthernets int      `json:"iol_static_ethernets"`
	Loopback           []string `json:"loopback"`
	Management         []string `json:"management"`
}

// NodeDefinitionMap is a map of node UUIDs to node definitions.
type NodeDefinitionMap map[UUID]NodeDefinition

// Supporting structs for OpenAPI schema compliance

// VMProperties represents VM property requirements.
type VMProperties struct {
	RAM          bool `json:"ram"`
	CPUs         bool `json:"cpus"`
	DataVolume   bool `json:"data_volume"`
	BootDiskSize bool `json:"boot_disk_size"`
	CPULimit     bool `json:"cpu_limit"`
}

// UsageEstimations represents resource usage estimates.
type UsageEstimations struct {
	CPUs int `json:"cpus"`
	RAM  int `json:"ram"`
	Disk int `json:"disk"`
}

// NodeParameters represents node-specific parameters.
type NodeParameters map[string]any

// DeviceNature represents the nature/type of a device.
type DeviceNature string

const (
	// DeviceNatureRouter represents a router device.
	DeviceNatureRouter DeviceNature = "router"
	// DeviceNatureSwitch represents a switch device.
	DeviceNatureSwitch DeviceNature = "switch"
	// DeviceNatureServer represents a server device.
	DeviceNatureServer DeviceNature = "server"
	// DeviceNatureHost represents a host device.
	DeviceNatureHost DeviceNature = "host"
	// DeviceNatureCloud represents a cloud device.
	DeviceNatureCloud DeviceNature = "cloud"
	// DeviceNatureFirewall represents a firewall device.
	DeviceNatureFirewall DeviceNature = "firewall"
	// DeviceNatureExternalConnector represents an external connector.
	DeviceNatureExternalConnector DeviceNature = "external_connector"
)

// NodeDefIcons represents the icon type for a node definition.
type NodeDefIcons string

const (
	// NodeDefIconsRouter represents a router icon.
	NodeDefIconsRouter NodeDefIcons = "router"
	// NodeDefIconsSwitch represents a switch icon.
	NodeDefIconsSwitch NodeDefIcons = "switch"
	// NodeDefIconsServer represents a server icon.
	NodeDefIconsServer NodeDefIcons = "server"
	// NodeDefIconsHost represents a host icon.
	NodeDefIconsHost NodeDefIcons = "host"
	// NodeDefIconsCloud represents a cloud icon.
	NodeDefIconsCloud NodeDefIcons = "cloud"
	// NodeDefIconsFirewall represents a firewall icon.
	NodeDefIconsFirewall NodeDefIcons = "firewall"
	// NodeDefIconsAccessPoint represents an access point icon.
	NodeDefIconsAccessPoint NodeDefIcons = "access_point"
	// NodeDefIconsWl represents a wireless icon.
	NodeDefIconsWl NodeDefIcons = "wl"
)

// LinuxNativeSimulation represents Linux native simulation settings.
type LinuxNativeSimulation struct {
	LibvirtDomainDriver string         `json:"libvirt_domain_driver"`
	Driver              string         `json:"driver"`
	DiskDriver          string         `json:"disk_driver"`
	EFIBoot             bool           `json:"efi_boot"`
	EFICode             string         `json:"efi_code"`
	EFIVars             string         `json:"efi_vars"`
	MachineType         string         `json:"machine_type"`
	RAM                 int            `json:"ram"`
	CPUs                int            `json:"cpus"`
	CPULimit            int            `json:"cpu_limit"`
	CPUModel            string         `json:"cpu_model"`
	NICDriver           string         `json:"nic_driver"`
	DataVolume          int            `json:"data_volume"`
	BootDiskSize        int            `json:"boot_disk_size"`
	Video               map[string]any `json:"video"`
	EnableRNG           bool           `json:"enable_rng"`
	EnableTPM           bool           `json:"enable_tpm"`
}

// NodeDefinitionBoot represents boot configuration for a node definition.
type NodeDefinitionBoot struct {
	Timeout   int      `json:"timeout"`
	Completed []string `json:"completed"`
	UsesRegex bool     `json:"uses_regex"`
}

// NodeDefinitionGeneral represents general configuration for a node definition.
type NodeDefinitionGeneral struct {
	Nature      DeviceNature `json:"nature"`
	Description string       `json:"description"`
	ReadOnly    bool         `json:"read_only"`
}

// GeneratorConfig represents generator configuration.
type GeneratorConfig struct {
	Driver string `json:"driver"`
}

// ProvisioningConfig represents provisioning configuration.
type ProvisioningConfig struct {
	Files      []map[string]any `json:"files"`
	MediaType  string           `json:"media_type"`
	VolumeName string           `json:"volume_name"`
}

// NodeDefinitionConfiguration represents the configuration for a node definition.
type NodeDefinitionConfiguration struct {
	Generator    GeneratorConfig    `json:"generator"`
	Provisioning ProvisioningConfig `json:"provisioning"`
}

// NodeDefinitionInherited represents inherited properties for a node definition.
type NodeDefinitionInherited struct {
	Image VMProperties `json:"image"`
	Node  VMProperties `json:"node"`
}

// NodeDefinitionPyats represents pyATS configuration for a node definition.
type NodeDefinitionPyats struct {
	OS                   string  `json:"os"`
	Series               string  `json:"series"`
	Model                string  `json:"model"`
	UseInTestbed         bool    `json:"use_in_testbed"`
	Username             *string `json:"username"`
	Password             *string `json:"password"`
	ConfigExtractCommand *string `json:"config_extract_command"`
}

// NodeDefinitionUI represents UI configuration for a node definition.
type NodeDefinitionUI struct {
	LabelPrefix         string       `json:"label_prefix"`
	Icon                NodeDefIcons `json:"icon"`
	Label               string       `json:"label"`
	Visible             bool         `json:"visible"`
	Group               string       `json:"group"`
	Description         string       `json:"description"`
	HasConfiguration    bool         `json:"has_configuration"`
	ShowRAM             bool         `json:"show_ram"`
	ShowCPUs            bool         `json:"show_cpus"`
	ShowCPULimit        bool         `json:"show_cpu_limit"`
	ShowDataVolume      bool         `json:"show_data_volume"`
	ShowBootDiskSize    bool         `json:"show_boot_disk_size"`
	HasConfigExtraction bool         `json:"has_config_extraction"`
}

// HasVNC returns true if the node definition supports VNC.
func (nd NodeDefinition) HasVNC() bool {
	return nd.Sim.VNC
}

// HasSerial returns true if the node definition has serial ports.
func (nd NodeDefinition) HasSerial() bool {
	return nd.Device.Interfaces.SerialPorts > 0
}

// SerialPorts returns the number of serial ports.
func (nd NodeDefinition) SerialPorts() int {
	return nd.Device.Interfaces.SerialPorts
}
