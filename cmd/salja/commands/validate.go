package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/gongahkia/salja/internal/model"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate and preview a calendar/task file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			if format == "" {
				format = DetectFormat(filePath)
			}

			ctx := context.Background()
			collection, err := ReadInput(ctx, filePath, format, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
				return err
			}

			col := collection
			fmt.Printf("File:     %s\n", filePath)
			fmt.Printf("Format:   %s\n", format)
			fmt.Printf("Items:    %d\n", len(col.Items))

			events, tasks, journals := 0, 0, 0
			withDue, withStart, withRecurrence, withTags := 0, 0, 0, 0
			for _, item := range col.Items {
				switch item.ItemType {
				case model.ItemTypeEvent:
					events++
				case model.ItemTypeTask:
					tasks++
				case model.ItemTypeJournal:
					journals++
				}
				if item.DueDate != nil {
					withDue++
				}
				if item.StartTime != nil {
					withStart++
				}
				if item.Recurrence != nil {
					withRecurrence++
				}
				if len(item.Tags) > 0 {
					withTags++
				}
			}

			fmt.Printf("Events:   %d\n", events)
			fmt.Printf("Tasks:    %d\n", tasks)
			fmt.Printf("Journals: %d\n", journals)
			fmt.Println("--- Field Coverage ---")
			fmt.Printf("With start time:  %d\n", withStart)
			fmt.Printf("With due date:    %d\n", withDue)
			fmt.Printf("With recurrence:  %d\n", withRecurrence)
			fmt.Printf("With tags:        %d\n", withTags)
			fmt.Println("✓ File is valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "Override format detection")
	return cmd
}
