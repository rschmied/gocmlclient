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

type NodeDefinitionMap map[UUID]NodeDefinition

// Supporting structs for OpenAPI schema compliance

type VMProperties struct {
	RAM          bool `json:"ram"`
	CPUs         bool `json:"cpus"`
	DataVolume   bool `json:"data_volume"`
	BootDiskSize bool `json:"boot_disk_size"`
	CPULimit     bool `json:"cpu_limit"`
}

type UsageEstimations struct {
	CPUs int `json:"cpus"`
	RAM  int `json:"ram"`
	Disk int `json:"disk"`
}

type NodeParameters map[string]any

type DeviceNature string

const (
	DeviceNatureRouter            DeviceNature = "router"
	DeviceNatureSwitch            DeviceNature = "switch"
	DeviceNatureServer            DeviceNature = "server"
	DeviceNatureHost              DeviceNature = "host"
	DeviceNatureCloud             DeviceNature = "cloud"
	DeviceNatureFirewall          DeviceNature = "firewall"
	DeviceNatureExternalConnector DeviceNature = "external_connector"
)

type NodeDefIcons string

const (
	NodeDefIconsRouter      NodeDefIcons = "router"
	NodeDefIconsSwitch      NodeDefIcons = "switch"
	NodeDefIconsServer      NodeDefIcons = "server"
	NodeDefIconsHost        NodeDefIcons = "host"
	NodeDefIconsCloud       NodeDefIcons = "cloud"
	NodeDefIconsFirewall    NodeDefIcons = "firewall"
	NodeDefIconsAccessPoint NodeDefIcons = "access_point"
	NodeDefIconsWl          NodeDefIcons = "wl"
)

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

type NodeDefinitionBoot struct {
	Timeout   int      `json:"timeout"`
	Completed []string `json:"completed"`
	UsesRegex bool     `json:"uses_regex"`
}

type NodeDefinitionGeneral struct {
	Nature      DeviceNature `json:"nature"`
	Description string       `json:"description"`
	ReadOnly    bool         `json:"read_only"`
}

type GeneratorConfig struct {
	Driver string `json:"driver"`
}

type ProvisioningConfig struct {
	Files      []map[string]any `json:"files"`
	MediaType  string           `json:"media_type"`
	VolumeName string           `json:"volume_name"`
}

type NodeDefinitionConfiguration struct {
	Generator    GeneratorConfig    `json:"generator"`
	Provisioning ProvisioningConfig `json:"provisioning"`
}

type NodeDefinitionInherited struct {
	Image VMProperties `json:"image"`
	Node  VMProperties `json:"node"`
}

type NodeDefinitionPyats struct {
	OS                   string  `json:"os"`
	Series               string  `json:"series"`
	Model                string  `json:"model"`
	UseInTestbed         bool    `json:"use_in_testbed"`
	Username             *string `json:"username"`
	Password             *string `json:"password"`
	ConfigExtractCommand *string `json:"config_extract_command"`
}

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
