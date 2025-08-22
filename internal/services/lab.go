// Example: internal/services/lab.go
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// LabService provides lab-related operations
type LabService struct {
	apiClient *api.Client
}

// NewLabService creates a new lab service
func NewLabService(apiClient *api.Client) *LabService {
	return &LabService{
		apiClient: apiClient,
	}
}

// Create creates a new lab on the controller
func (s *LabService) Create(ctx context.Context, lab *models.Lab) (*models.Lab, error) {
	// Use the alias type for API communication (handles inconsistent field names)
	postData := labCreateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
	}

	var response labResponse
	err := s.apiClient.PostJSON(ctx, "labs", nil, postData, &response)
	if err != nil {
		return nil, fmt.Errorf("create lab: %w", err)
	}

	// Update the lab with the ID from the response
	lab.ID = response.ID

	// Update with full data (handles groups, owner, etc.)
	return s.Update(ctx, lab)
}

// Get retrieves a lab by ID
func (s *LabService) Get(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	endpoint := fmt.Sprintf("labs/%s", id)

	var response labResponse
	err := s.apiClient.GetJSON(ctx, endpoint, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("get lab %s: %w", id, err)
	}

	lab := response.ToLab()

	if deep {
		// Fetch additional data (nodes, links, etc.)
		return s.fillLabData(ctx, lab)
	}

	return lab, nil
}

// Update updates a lab's metadata
func (s *LabService) Update(ctx context.Context, lab *models.Lab) (*models.Lab, error) {
	endpoint := fmt.Sprintf("labs/%s", lab.ID)

	updateData := labUpdateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
		Groups:      lab.Groups,
	}

	var response labResponse
	err := s.apiClient.PatchJSON(ctx, endpoint, updateData, &response)
	if err != nil {
		return nil, fmt.Errorf("update lab %s: %w", lab.ID, err)
	}

	return response.ToLab(), nil
}

// Start starts all nodes in a lab
func (s *LabService) Start(ctx context.Context, id string) error {
	endpoint := fmt.Sprintf("labs/%s/start", id)
	err := s.apiClient.PutJSON(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("start lab %s: %w", id, err)
	}
	return nil
}

// Stop stops all nodes in a lab
func (s *LabService) Stop(ctx context.Context, id string) error {
	endpoint := fmt.Sprintf("labs/%s/stop", id)
	err := s.apiClient.PutJSON(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("stop lab %s: %w", id, err)
	}
	return nil
}

// Import imports a lab from YAML topology
func (s *LabService) Import(ctx context.Context, topology string) (*models.Lab, error) {
	topoReader := strings.NewReader(topology)

	var importResponse struct {
		ID       string   `json:"id"`
		Warnings []string `json:"warnings"`
	}

	err := s.apiClient.PostJSON(ctx, "import", nil, topoReader, &importResponse)
	if err != nil {
		return nil, fmt.Errorf("import lab: %w", err)
	}

	if len(importResponse.Warnings) > 0 {
		slog.Warn("Lab import completed with warnings", "warnings", importResponse.Warnings)
	}

	// Fetch the imported lab with full data
	return s.Get(ctx, importResponse.ID, true)
}

// HasConverged checks if all nodes in the lab have converged (are in BOOTED state)
func (s *LabService) HasConverged(ctx context.Context, id string) (bool, error) {
	endpoint := fmt.Sprintf("labs/%s/check_if_converged", id)

	var converged bool
	err := s.apiClient.GetJSON(ctx, endpoint, nil, &converged)
	if err != nil {
		return false, fmt.Errorf("check lab convergence %s: %w", id, err)
	}

	return converged, nil
}

// fillLabData fetches additional lab data for deep queries
func (s *LabService) fillLabData(ctx context.Context, lab *models.Lab) (*models.Lab, error) {
	// This would use errgroup to fetch nodes, links, L3 info concurrently
	// Similar to your current labFill function but using the api.Client
	// Implementation details omitted for brevity
	return lab, nil
}

// API request/response types (internal to the service)
type labCreateRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type labUpdateRequest struct {
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Notes       string             `json:"notes,omitempty"`
	Groups      []*models.LabGroup `json:"groups,omitempty"`
}

type labResponse struct {
	ID          string             `json:"id"`
	State       string             `json:"state"`
	Created     string             `json:"created"`
	Modified    string             `json:"modified"`
	Title       string             `json:"lab_title"`
	Description string             `json:"lab_description"`
	Notes       string             `json:"lab_notes"`
	OwnerID     string             `json:"owner"`
	NodeCount   int                `json:"node_count"`
	LinkCount   int                `json:"link_count"`
	Groups      []*models.LabGroup `json:"groups"`
}

func (r *labResponse) ToLab() *models.Lab {
	return &models.Lab{
		ID:          r.ID,
		State:       r.State,
		Created:     r.Created,
		Modified:    r.Modified,
		Title:       r.Title,
		Description: r.Description,
		Notes:       r.Notes,
		Owner:       &models.User{ID: r.OwnerID},
		NodeCount:   r.NodeCount,
		LinkCount:   r.LinkCount,
		Groups:      r.Groups,
		Nodes:       make(models.NodeMap),
		Links:       []*models.Link{},
	}
}

// ====================================================================
// Example: How to initialize the API client (internal/api/example.go)
// ====================================================================

// func ExampleAPIClientUsage() {
// 	// Create transport with custom configuration
// 	transportConfig := api.DefaultTransportConfig()
// 	transportConfig.InsecureSkipVerify = true // for development
// 	transport := api.NewTransport(transportConfig)
//
// 	// Create HTTP client
// 	httpClient := api.NewHTTPClient(transport, 15*time.Second)
//
// 	// Create middlewares
// 	middlewares := []api.Middleware{
// 		api.LoggingMiddleware(slog.Default()),
// 		api.LogRequestBodyMiddleware(slog.Default()),
// 		api.RetryMiddleware(api.DefaultRetryPolicy()),
// 		api.UserAgentMiddleware("gocmlclient/1.0"),
// 	}
//
// 	// Create API client
// 	apiClient := api.New("https://cml-controller.example.com", api.Options{
// 		HTTPClient:  httpClient,
// 		Middlewares: middlewares,
// 	})
//
// 	// Use the API client in services
// 	labService := NewLabService(apiClient)
//
// 	// The service provides a clean interface
// 	ctx := context.Background()
// 	lab, err := labService.Get(ctx, "lab-uuid", true)
// 	if err != nil {
// 		slog.Error("Failed to get lab", "error", err)
// 		return
// 	}
//
// 	slog.Info("Retrieved lab", "title", lab.Title, "nodes", len(lab.Nodes))
// }
