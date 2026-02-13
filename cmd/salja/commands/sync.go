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

if token.IsExpired() {
return fmt.Errorf("token for %s is expired; run: salja auth login %s", to, to)
}

switch to {
case "google":
return pushToGoogle(ctx, token, collection, dryRun, apiTimeout)
case "microsoft":
return pushToMicrosoft(ctx, token, collection, dryRun, apiTimeout)
default:
return fmt.Errorf("unsupported target %q; supported: google, microsoft", to)
}
},
}

cmd.Flags().StringVar(&to, "to", "", "Target service: google, microsoft")
cmd.MarkFlagRequired("to")
cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without making API calls")
return cmd
}

func newSyncPullCmd() *cobra.Command {
var from, output string

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

if token.IsExpired() {
return fmt.Errorf("token for %s is expired; run: salja auth login %s", from, from)
}

cfg, _ := config.Load()
apiTimeout := 30 * time.Second
if cfg != nil && cfg.APITimeoutSeconds > 0 {
apiTimeout = time.Duration(cfg.APITimeoutSeconds) * time.Second
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

now := time.Now()
startTime := now.AddDate(0, -1, 0)
endTime := now.AddDate(0, 3, 0)

var collection *model.CalendarCollection
switch from {
case "google":
collection, err = pullFromGoogle(ctx, token, startTime, endTime, apiTimeout)
case "microsoft":
collection, err = pullFromMicrosoft(ctx, token, startTime, endTime, apiTimeout)
default:
return fmt.Errorf("unsupported source %q; supported: google, microsoft", from)
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

cmd.Flags().StringVar(&from, "from", "", "Source service: google, microsoft")
cmd.MarkFlagRequired("from")
cmd.Flags().StringVar(&output, "output", "", "Output file path")
cmd.MarkFlagRequired("output")
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
return nil, fmt.Errorf("Google Calendar API error: %w", err)
}

collection := &model.CalendarCollection{
Items:     make([]model.CalendarItem, 0, len(events)),
SourceApp: "google",
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
return nil, fmt.Errorf("Microsoft Graph API error: %w", err)
}

collection := &model.CalendarCollection{
Items:     make([]model.CalendarItem, 0, len(events)),
SourceApp: "microsoft",
ExportDate: time.Now(),
}
for _, event := range events {
collection.Items = append(collection.Items, api.MSGraphToCalendarItem(event))
}
return collection, nil
}
