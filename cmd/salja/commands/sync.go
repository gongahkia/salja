package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gongahkia/salja/internal/api"
	"github.com/gongahkia/salja/internal/config"
	"github.com/gongahkia/salja/internal/model"
	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync calendar data with cloud services",
	}

	cmd.AddCommand(newSyncPushCmd())
	cmd.AddCommand(newSyncPullCmd())
	return cmd
}

func ensureTokenValid(ctx context.Context, store api.TokenStorer, service string, token *api.Token, cfg *config.Config) (*api.Token, error) {
	if !token.IsExpired() {
		return token, nil
	}
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("token for %s is expired; run: salja auth login %s", service, service)
	}
	var pkceConfig api.PKCEConfig
	switch service {
	case "google":
		if cfg != nil {
			pkceConfig = api.PKCEConfig{ClientID: cfg.API.Google.ClientID, TokenURL: "https://oauth2.googleapis.com/token"}
		}
	case "microsoft":
		if cfg != nil {
			pkceConfig = api.PKCEConfig{ClientID: cfg.API.Microsoft.ClientID, TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token"}
		}
	case "todoist":
		if cfg != nil {
			pkceConfig = api.PKCEConfig{ClientID: cfg.API.Todoist.ClientID, TokenURL: "https://todoist.com/oauth/access_token"}
		}
	case "ticktick":
		if cfg != nil {
			pkceConfig = api.PKCEConfig{ClientID: cfg.API.TickTick.ClientID, TokenURL: "https://ticktick.com/oauth/token"}
		}
	default:
		return nil, fmt.Errorf("token for %s is expired; run: salja auth login %s", service, service)
	}
	flow := api.NewPKCEFlow(pkceConfig)
	newToken, err := flow.RefreshAccessToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh %s token: %w", service, err)
	}
	if err := store.Set(service, newToken); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to persist refreshed token: %v\n", err)
	}
	return newToken, nil
}

func newSyncPushCmd() *cobra.Command {
	var to string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "push <file>",
		Short: "Push local file items to a cloud service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			format := DetectFormat(filePath)

			cfg, _ := config.Load()
			apiTimeout := 30 * time.Second
			if cfg != nil && cfg.APITimeoutSeconds > 0 {
				apiTimeout = time.Duration(cfg.APITimeoutSeconds) * time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			collection, err := ReadInput(ctx, filePath, format, nil)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			store, err := api.DefaultSecureStore()
			if err != nil {
				return err
			}

			token, err := store.Get(to)
			if err != nil {
				return err
			}

			token, err = ensureTokenValid(ctx, store, to, token, cfg)
			if err != nil {
				return err
			}

			switch to {
			case "google":
				return pushToGoogle(ctx, token, collection, dryRun, apiTimeout)
			case "microsoft":
				return pushToMicrosoft(ctx, token, collection, dryRun, apiTimeout)
			case "todoist":
				return pushToTodoist(ctx, token, collection, dryRun, apiTimeout)
			case "ticktick":
				return pushToTickTick(ctx, token, collection, dryRun, apiTimeout)
			case "notion":
				return pushToNotion(ctx, token, collection, dryRun, apiTimeout)
			default:
				return fmt.Errorf("unsupported target %q; supported: google, microsoft, todoist, ticktick, notion", to)
			}
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Target service: google, microsoft")
	_ = cmd.MarkFlagRequired("to")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without making API calls")
	return cmd
}

func newSyncPullCmd() *cobra.Command {
	var from, output, startFlag, endFlag string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull events from a cloud service to a local file",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := api.DefaultSecureStore()
			if err != nil {
				return err
			}

			token, err := store.Get(from)
			if err != nil {
				return err
			}

			cfg, _ := config.Load()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			token, err = ensureTokenValid(ctx, store, from, token, cfg)
			if err != nil {
				return err
			}

			apiTimeout := 30 * time.Second
			if cfg != nil && cfg.APITimeoutSeconds > 0 {
				apiTimeout = time.Duration(cfg.APITimeoutSeconds) * time.Second
			}

			now := time.Now()
			startTime := now.AddDate(0, -1, 0)
			endTime := now.AddDate(0, 3, 0)

			if startFlag != "" {
				if t, err := time.Parse("2006-01-02", startFlag); err == nil {
					startTime = t
				} else {
					return fmt.Errorf("invalid --start date %q; use YYYY-MM-DD", startFlag)
				}
			}
			if endFlag != "" {
				if t, err := time.Parse("2006-01-02", endFlag); err == nil {
					endTime = t
				} else {
					return fmt.Errorf("invalid --end date %q; use YYYY-MM-DD", endFlag)
				}
			}

			var collection *model.CalendarCollection
			switch from {
			case "google":
				collection, err = pullFromGoogle(ctx, token, startTime, endTime, apiTimeout)
			case "microsoft":
				collection, err = pullFromMicrosoft(ctx, token, startTime, endTime, apiTimeout)
			case "todoist":
				collection, err = pullFromTodoist(ctx, token, apiTimeout)
			case "ticktick":
				collection, err = pullFromTickTick(ctx, token, apiTimeout)
			case "notion":
				collection, err = pullFromNotion(ctx, token, apiTimeout)
			default:
				return fmt.Errorf("unsupported source %q; supported: google, microsoft, todoist, ticktick, notion", from)
			}
			if err != nil {
				return err
			}

			outFormat := DetectFormat(output)
			if err := WriteOutput(ctx, collection, output, outFormat); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Pulled %d items from %s to %s\n", len(collection.Items), from, output)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Source service: google, microsoft, todoist, ticktick, notion")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().StringVar(&output, "output", "", "Output file path")
	_ = cmd.MarkFlagRequired("output")
	cmd.Flags().StringVar(&startFlag, "start", "", "Start date for pull range (YYYY-MM-DD, default: -1 month)")
	cmd.Flags().StringVar(&endFlag, "end", "", "End date for pull range (YYYY-MM-DD, default: +3 months)")
	return cmd
}

func pushToGoogle(ctx context.Context, token *api.Token, collection *model.CalendarCollection, dryRun bool, timeout time.Duration) error {
	client := api.NewGCalClientWithTimeout(token, timeout)
	created := 0
	// Google Calendar API quota: 10 QPS for calendar.events.insert
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for _, item := range collection.Items {
		event := api.CalendarItemToGCal(item)
		if dryRun {
			fmt.Printf("  [dry-run] would create: %s\n", event.Summary)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		_, err := client.InsertEvent(ctx, "primary", &event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %s (%v)\n", item.Title, err)
			continue
		}
		created++
	}
	if !dryRun {
		fmt.Fprintf(os.Stderr, "✓ Created %d/%d events in Google Calendar\n", created, len(collection.Items))
	}
	return nil
}

func pushToMicrosoft(ctx context.Context, token *api.Token, collection *model.CalendarCollection, dryRun bool, timeout time.Duration) error {
	client := api.NewMSGraphClientWithTimeout(token, timeout)
	created := 0
	// Microsoft Graph rate limit: ~4 requests per second
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for _, item := range collection.Items {
		event := api.CalendarItemToMSGraph(item)
		if dryRun {
			fmt.Printf("  [dry-run] would create: %s\n", event.Subject)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		_, err := client.CreateEvent(ctx, &event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %s (%v)\n", item.Title, err)
			continue
		}
		created++
	}
	if !dryRun {
		fmt.Fprintf(os.Stderr, "✓ Created %d/%d events in Microsoft Outlook\n", created, len(collection.Items))
	}
	return nil
}

func pullFromGoogle(ctx context.Context, token *api.Token, startTime, endTime time.Time, timeout time.Duration) (*model.CalendarCollection, error) {
	client := api.NewGCalClientWithTimeout(token, timeout)
	events, err := client.ListEvents(ctx, "primary", startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("google calendar API error: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:      make([]model.CalendarItem, 0, len(events)),
		SourceApp:  "google",
		ExportDate: time.Now(),
	}
	for _, event := range events {
		collection.Items = append(collection.Items, api.GCalToCalendarItem(event))
	}
	return collection, nil
}

func pullFromMicrosoft(ctx context.Context, token *api.Token, startTime, endTime time.Time, timeout time.Duration) (*model.CalendarCollection, error) {
	client := api.NewMSGraphClientWithTimeout(token, timeout)
	events, err := client.ListEvents(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("microsoft graph API error: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:      make([]model.CalendarItem, 0, len(events)),
		SourceApp:  "microsoft",
		ExportDate: time.Now(),
	}
	for _, event := range events {
		collection.Items = append(collection.Items, api.MSGraphToCalendarItem(event))
	}
	return collection, nil
}

func pushToTodoist(ctx context.Context, token *api.Token, collection *model.CalendarCollection, dryRun bool, timeout time.Duration) error {
	client := api.NewTodoistClientWithTimeout(token, timeout)
	created := 0
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for _, item := range collection.Items {
		task := api.CalendarItemToTodoist(item, "")
		if dryRun {
			fmt.Printf("  [dry-run] would create: %s\n", task.Content)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		_, err := client.CreateTask(ctx, &task)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %s (%v)\n", item.Title, err)
			continue
		}
		created++
	}
	if !dryRun {
		fmt.Fprintf(os.Stderr, "✓ Created %d/%d tasks in Todoist\n", created, len(collection.Items))
	}
	return nil
}

func pullFromTodoist(ctx context.Context, token *api.Token, timeout time.Duration) (*model.CalendarCollection, error) {
	client := api.NewTodoistClientWithTimeout(token, timeout)
	tasks, err := client.GetTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("todoist API error: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:      make([]model.CalendarItem, 0, len(tasks)),
		SourceApp:  "todoist",
		ExportDate: time.Now(),
	}
	for _, task := range tasks {
		collection.Items = append(collection.Items, api.TodoistToCalendarItem(task))
	}
	return collection, nil
}

func pushToTickTick(ctx context.Context, token *api.Token, collection *model.CalendarCollection, dryRun bool, timeout time.Duration) error {
	client := api.NewTickTickClientWithTimeout(token, timeout)

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list TickTick projects: %w", err)
	}
	if len(projects) == 0 {
		return fmt.Errorf("no TickTick projects found")
	}

	fmt.Fprintln(os.Stderr, "Select a TickTick project:")
	for i, p := range projects {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, p.Name)
	}
	fmt.Fprint(os.Stderr, "> ")
	var choice int
	if _, err := fmt.Fscan(os.Stdin, &choice); err != nil || choice < 1 || choice > len(projects) {
		choice = 1
	}
	projectID := projects[choice-1].ID

	created := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for _, item := range collection.Items {
		task := api.CalendarItemToTickTick(item, projectID)
		if dryRun {
			fmt.Printf("  [dry-run] would create: %s\n", task.Title)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		_, err := client.CreateTask(ctx, &task)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %s (%v)\n", item.Title, err)
			continue
		}
		created++
	}
	if !dryRun {
		fmt.Fprintf(os.Stderr, "✓ Created %d/%d tasks in TickTick\n", created, len(collection.Items))
	}
	return nil
}

func pullFromTickTick(ctx context.Context, token *api.Token, timeout time.Duration) (*model.CalendarCollection, error) {
	client := api.NewTickTickClientWithTimeout(token, timeout)

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list TickTick projects: %w", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("no TickTick projects found")
	}

	fmt.Fprintln(os.Stderr, "Select a TickTick project to pull from:")
	for i, p := range projects {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, p.Name)
	}
	fmt.Fprint(os.Stderr, "> ")
	var choice int
	if _, err := fmt.Fscan(os.Stdin, &choice); err != nil || choice < 1 || choice > len(projects) {
		choice = 1
	}

	tasks, err := client.ListTasks(ctx, projects[choice-1].ID)
	if err != nil {
		return nil, fmt.Errorf("TickTick API error: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:      make([]model.CalendarItem, 0, len(tasks)),
		SourceApp:  "ticktick",
		ExportDate: time.Now(),
	}
	for _, task := range tasks {
		collection.Items = append(collection.Items, api.TickTickToCalendarItem(task))
	}
	return collection, nil
}

func pushToNotion(ctx context.Context, token *api.Token, collection *model.CalendarCollection, dryRun bool, timeout time.Duration) error {
	client := api.NewNotionClientWithTimeout(token.AccessToken, timeout)
	pm := api.DefaultNotionPropertyMap()

	fmt.Fprint(os.Stderr, "Enter Notion database ID: ")
	var databaseID string
	_, _ = fmt.Fscanln(os.Stdin, &databaseID)
	if databaseID == "" {
		return fmt.Errorf("database ID cannot be empty")
	}

	created := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for _, item := range collection.Items {
		page := api.CalendarItemToNotion(item, pm)
		if dryRun {
			fmt.Printf("  [dry-run] would create: %s\n", item.Title)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		_, err := client.CreatePage(ctx, databaseID, &page)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %s (%v)\n", item.Title, err)
			continue
		}
		created++
	}
	if !dryRun {
		fmt.Fprintf(os.Stderr, "✓ Created %d/%d pages in Notion\n", created, len(collection.Items))
	}
	return nil
}

func pullFromNotion(ctx context.Context, token *api.Token, timeout time.Duration) (*model.CalendarCollection, error) {
	client := api.NewNotionClientWithTimeout(token.AccessToken, timeout)
	pm := api.DefaultNotionPropertyMap()

	fmt.Fprint(os.Stderr, "Enter Notion database ID: ")
	var databaseID string
	_, _ = fmt.Fscanln(os.Stdin, &databaseID)
	if databaseID == "" {
		return nil, fmt.Errorf("database ID cannot be empty")
	}

	collection := &model.CalendarCollection{
		Items:      []model.CalendarItem{},
		SourceApp:  "notion",
		ExportDate: time.Now(),
	}

	cursor := ""
	for {
		result, err := client.QueryDatabase(ctx, databaseID, cursor)
		if err != nil {
			return nil, fmt.Errorf("notion API error: %w", err)
		}
		for _, page := range result.Results {
			collection.Items = append(collection.Items, api.NotionToCalendarItem(page, pm))
		}
		if !result.HasMore {
			break
		}
		cursor = result.NextCursor
	}

	return collection, nil
}
