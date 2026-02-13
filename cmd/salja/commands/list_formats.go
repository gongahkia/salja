package commands

import (
	"fmt"
	"sort"

	"github.com/gongahkia/salja/internal/registry"
	_ "github.com/gongahkia/salja/internal/registry" // ensure format registration
	"github.com/spf13/cobra"
)

func NewListFormatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-formats",
		Short: "List all supported formats with capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Format       | Events | Tasks | Recurring | Subtasks")
			fmt.Println("-------------|--------|-------|-----------|--------")

			allFormats := registry.AllFormats()
			names := make([]string, 0, len(allFormats))
			for name := range allFormats {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				caps := allFormats[name].Capabilities
				fmt.Printf("%-13s|  %s   |  %s  |    %s    |  %s\n",
					name,
					checkmark(caps.SupportsEvents),
					checkmark(caps.SupportsTasks),
					checkmark(caps.SupportsRecurrence),
					checkmark(caps.SupportsSubtasks),
				)
			}
		},
	}
}

func checkmark(v bool) string {
	if v {
		return " ✓ "
	}
	return "   "
}
