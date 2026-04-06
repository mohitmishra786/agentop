package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/agentop-dev/agentop/internal/pricing"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show effective config and pricing snapshot",
	RunE:  runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	db := pricing.GetDB()

	fmt.Println("agentop configuration")
	fmt.Println()
	fmt.Printf("Claude data dir: %s\n", claudeDir)
	fmt.Printf("Pricing version: %s\n", db.Version)
	fmt.Printf("Models in pricing table: %d\n", len(db.Models))
	fmt.Println()

	fmt.Println("Model pricing (USD per 1M tokens):")
	for name, p := range db.Models {
		fmt.Printf("  %-25s  in: $%-7.2f  out: $%-7.2f  cc: $%-7.2f  cr: $%-7.2f\n",
			name, p.Input, p.Output, p.CacheCreate, p.CacheRead)
	}

	fmt.Println()
	fmt.Printf("since: %s\n", since)
	if until != "" {
		fmt.Printf("until: %s\n", until)
	}
	if project != "" {
		fmt.Printf("project: %s\n", project)
	}
	if model != "" {
		fmt.Printf("model: %s\n", model)
	}
	fmt.Printf("watch: %v\n", watch)
	fmt.Printf("refresh: %ds\n", refresh)
	fmt.Printf("compact: %v\n", compact)
	fmt.Printf("no-color: %v\n", noColor)

	return nil
}
