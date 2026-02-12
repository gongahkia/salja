package api

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"strings"
"time"

"github.com/gongahkia/calendar-converter/internal/model"
)

const (
todoistRESTURL = "https://api.todoist.com/rest/v2"
todoistSyncURL = "https://api.todoist.com/sync/v9"
)

// TodoistClient is a REST API client for Todoist.
type TodoistClient struct {
token      *Token
httpClient *http.Client
}

func NewTodoistClient(token *Token) *TodoistClient {
return &TodoistClient{
token:      token,
httpClient: &http.Client{Timeout: 30 * time.Second},
}
}

func (c *TodoistClient) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
var reqBody io.Reader
if body != nil {
data, err := json.Marshal(body)
if err != nil {
return nil, 0, err
}
reqBody = bytes.NewReader(data)
}

req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
if err != nil {
return nil, 0, err
}
req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
req.Header.Set("Content-Type", "application/json")

resp, err := c.httpClient.Do(req)
if err != nil {
return nil, 0, err
}
defer resp.Body.Close()

respBody, err := io.ReadAll(resp.Body)
return respBody, resp.StatusCode, err
}

// TodoistTask represents a Todoist task.
type TodoistTask struct {
ID          string       `json:"id"`
ProjectID   string       `json:"project_id"`
SectionID   string       `json:"section_id,omitempty"`
ParentID    string       `json:"parent_id,omitempty"`
Content     string       `json:"content"`
Description string       `json:"description,omitempty"`
Priority    int          `json:"priority"`
Labels      []string     `json:"labels,omitempty"`
Due         *TodoistDue  `json:"due,omitempty"`
IsCompleted bool         `json:"is_completed"`
Order       int          `json:"order"`
}

type TodoistDue struct {
Date      string `json:"date"`
String    string `json:"string,omitempty"`
Datetime  string `json:"datetime,omitempty"`
Timezone  string `json:"timezone,omitempty"`
IsRecurring bool `json:"is_recurring"`
}

type TodoistProject struct {
ID   string `json:"id"`
Name string `json:"name"`
}

func (c *TodoistClient) GetTasks(ctx context.Context) ([]TodoistTask, error) {
data, status, err := c.doRequest(ctx, "GET", todoistRESTURL+"/tasks", nil)
if err != nil {
return nil, err
}
if status != 200 {
return nil, fmt.Errorf("Todoist API error (HTTP %d): %s", status, data)
}
var tasks []TodoistTask
return tasks, json.Unmarshal(data, &tasks)
}

func (c *TodoistClient) CreateTask(ctx context.Context, task *TodoistTask) (*TodoistTask, error) {
data, status, err := c.doRequest(ctx, "POST", todoistRESTURL+"/tasks", task)
if err != nil {
return nil, err
}
if status != 200 {
return nil, fmt.Errorf("Todoist API error (HTTP %d): %s", status, data)
}
var created TodoistTask
return &created, json.Unmarshal(data, &created)
}

func (c *TodoistClient) UpdateTask(ctx context.Context, taskID string, task *TodoistTask) error {
_, status, err := c.doRequest(ctx, "POST", todoistRESTURL+"/tasks/"+taskID, task)
if err != nil {
return err
}
if status != 204 {
return fmt.Errorf("Todoist update failed (HTTP %d)", status)
}
return nil
}

func (c *TodoistClient) CloseTask(ctx context.Context, taskID string) error {
_, status, err := c.doRequest(ctx, "POST", todoistRESTURL+"/tasks/"+taskID+"/close", nil)
if err != nil {
return err
}
if status != 204 {
return fmt.Errorf("Todoist close failed (HTTP %d)", status)
}
return nil
}

func (c *TodoistClient) GetProjects(ctx context.Context) ([]TodoistProject, error) {
data, status, err := c.doRequest(ctx, "GET", todoistRESTURL+"/projects", nil)
if err != nil {
return nil, err
}
if status != 200 {
return nil, fmt.Errorf("Todoist API error (HTTP %d): %s", status, data)
}
var projects []TodoistProject
return projects, json.Unmarshal(data, &projects)
}

// TodoistSyncClient handles incremental sync via sync tokens.
type TodoistSyncClient struct {
token      *Token
syncToken  string
httpClient *http.Client
}

func NewTodoistSyncClient(token *Token) *TodoistSyncClient {
return &TodoistSyncClient{
token:      token,
syncToken:  "*",
httpClient: &http.Client{Timeout: 30 * time.Second},
}
}

type TodoistSyncResponse struct {
SyncToken string          `json:"sync_token"`
Items     json.RawMessage `json:"items"`
Projects  json.RawMessage `json:"projects"`
}

func (c *TodoistSyncClient) Sync(ctx context.Context, resourceTypes []string) (*TodoistSyncResponse, error) {
body := map[string]interface{}{
"sync_token":     c.syncToken,
"resource_types": resourceTypes,
}
data, err := json.Marshal(body)
if err != nil {
return nil, err
}

req, err := http.NewRequestWithContext(ctx, "POST", todoistSyncURL+"/sync", bytes.NewReader(data))
if err != nil {
return nil, err
}
req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
req.Header.Set("Content-Type", "application/json")

resp, err := c.httpClient.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != 200 {
respBody, _ := io.ReadAll(resp.Body)
return nil, fmt.Errorf("Todoist Sync API error (HTTP %d): %s", resp.StatusCode, respBody)
}

var syncResp TodoistSyncResponse
if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
return nil, err
}

c.syncToken = syncResp.SyncToken
return &syncResp, nil
}

// TodoistToCalendarItem maps a Todoist task to the unified model.
func TodoistToCalendarItem(task TodoistTask) model.CalendarItem {
item := model.CalendarItem{
UID:         task.ID,
Title:       task.Content,
Description: task.Description,
ItemType:    model.ItemTypeTask,
Tags:        task.Labels,
}

// Todoist priority: 4=urgent, 3=high, 2=medium, 1=normal (inverted)
switch task.Priority {
case 4:
item.Priority = model.PriorityHighest
case 3:
item.Priority = model.PriorityHigh
case 2:
item.Priority = model.PriorityMedium
default:
item.Priority = model.PriorityLow
}

if task.IsCompleted {
item.Status = model.StatusCompleted
} else {
item.Status = model.StatusPending
}

if task.Due != nil {
if task.Due.Datetime != "" {
if t, err := time.Parse(time.RFC3339, task.Due.Datetime); err == nil {
item.DueDate = &t
}
} else if task.Due.Date != "" {
if t, err := time.Parse("2006-01-02", task.Due.Date); err == nil {
item.DueDate = &t
}
}
if task.Due.IsRecurring && task.Due.String != "" {
item.Description = strings.TrimSpace(item.Description + "\nRecurrence: " + task.Due.String)
}
}

return item
}

// CalendarItemToTodoist maps the unified model to a Todoist task.
func CalendarItemToTodoist(item model.CalendarItem, projectID string) TodoistTask {
task := TodoistTask{
ID:        item.UID,
ProjectID: projectID,
Content:   item.Title,
Description: item.Description,
Labels:    item.Tags,
}

switch item.Priority {
case model.PriorityHighest:
task.Priority = 4
case model.PriorityHigh:
task.Priority = 3
case model.PriorityMedium:
task.Priority = 2
default:
task.Priority = 1
}

if item.Status == model.StatusCompleted {
task.IsCompleted = true
}

if item.DueDate != nil {
task.Due = &TodoistDue{
Date: item.DueDate.Format("2006-01-02"),
}
}

return task
}
