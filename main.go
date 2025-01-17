package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-sdk-gen/generator/python"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: coze-sdk-gen <openapi.yaml>")
		os.Exit(1)
	}

	yamlPath := os.Args[1]
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Printf("read yaml file failed: %v\n", err)
		os.Exit(1)
	}

	generator := python.Generator{}
	files, err := generator.Generate(context.Background(), yamlContent)
	if err != nil {
		fmt.Printf("generate sdk failed: %v\n", err)
		os.Exit(1)
	}

	// Write files
	for moduleName, content := range files {
		err := os.WriteFile(fmt.Sprintf("generated_%s.py", moduleName), []byte(content), 0644)
		if err != nil {
			fmt.Printf("write file failed: %v\n", err)
			os.Exit(1)
		}
	}
}
