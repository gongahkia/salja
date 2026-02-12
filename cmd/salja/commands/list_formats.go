package commands

import (
"fmt"

"github.com/spf13/cobra"
)

func NewListFormatsCmd() *cobra.Command {
return &cobra.Command{
Use:   "list-formats",
Short: "List all supported formats with capabilities",
Run: func(cmd *cobra.Command, args []string) {
fmt.Println("Format       | Events | Tasks | Recurring | Subtasks")
fmt.Println("-------------|--------|-------|-----------|--------")
fmt.Println("ics          |   ✓    |   ✓   |     ✓     |        ")
fmt.Println("ticktick     |        |   ✓   |     ✓     |   ✓   ")
fmt.Println("todoist      |        |   ✓   |           |   ✓   ")
fmt.Println("gcal         |   ✓    |       |           |        ")
fmt.Println("outlook      |   ✓    |       |           |        ")
fmt.Println("notion       |        |   ✓   |           |        ")
fmt.Println("trello       |        |   ✓   |           |   ✓   ")
fmt.Println("asana        |        |   ✓   |           |        ")
fmt.Println("omnifocus    |        |   ✓   |           |   ✓   ")
fmt.Println("apple-cal    |   ✓    |       |     ✓     |        ")
fmt.Println("apple-remind |        |   ✓   |           |        ")
fmt.Println("sorted3      |   ✓    |   ✓   |           |        ")
},
}
}
