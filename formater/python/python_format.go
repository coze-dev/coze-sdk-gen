package python

import (
	"context"
	"fmt"
	"os/exec"
)

func Format(ctx context.Context, path string) error {
	// Run ruff format on the generated files
	ruffCmd := exec.Command("poetry", "run", "ruff", "format", ".")
	ruffCmd.Dir = path
	ruffOutput, err := ruffCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Warning: Failed to run ruff format: %v\nOutput: %s\n", err, ruffOutput)
	} else {
		fmt.Println("Successfully formatted code with ruff!")
	}

	return nil
}
