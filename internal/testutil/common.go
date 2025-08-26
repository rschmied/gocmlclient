package testutil

import (
	"encoding/json"
	"fmt"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func PrettyPrintError(err error) error {
	var outer models.ErrorResponse
	if json.Unmarshal([]byte(err.Error()), &outer) == nil {
		fmt.Printf("code: %d\n", outer.Code)

		var rawDescription string
		if json.Unmarshal(outer.Description, &rawDescription) == nil {
			var innerData map[string]any
			if json.Unmarshal([]byte(rawDescription), &innerData) == nil {
				prettyJSON, err := json.MarshalIndent(innerData, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to pretty-print inner JSON: %w", err)
				}
				fmt.Printf("desc:\n%s\n", string(prettyJSON))
			} else {
				fmt.Printf("desc: %s\n", rawDescription)
			}
		} else {
			fmt.Printf("desc: %s\n", string(outer.Description))
		}
	} else {
		fmt.Printf("received non-JSON error: %v\n", err)
	}

	return nil
}
