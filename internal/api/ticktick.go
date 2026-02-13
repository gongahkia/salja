package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/model"
)

const tickTickBaseURL = "https://api.ticktick.com/open/v1"

// TickTickClient is a REST API client for TickTick.
type TickTickClient struct {
	token      *Token
	httpClient *http.Client
	baseURL    string
}

func NewTickTickClient(token *Token) *TickTickClient {
	return &TickTickClient{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    tickTickBaseURL,
	}
}

func NewTickTickClientWithTimeout(token *Token, timeout time.Duration) *TickTickClient {
	return &TickTickClient{
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    tickTickBaseURL,
	}
}

func (c *TickTickClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
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
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

// TickTickTask represents a task from the TickTick API.
type TickTickTask struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"projectId"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Desc       string            `json:"desc"`
	StartDate  string            `json:"startDate,omitempty"`
	DueDate    string            `json:"dueDate,omitempty"`
	Priority   int               `json:"priority"`
	Status     int               `json:"status"`
	Tags       []string          `json:"tags,omitempty"`
	Items      []TickTickSubtask `json:"items,omitempty"`
	TimeZone   string            `json:"timeZone,omitempty"`
	RepeatFlag string            `json:"repeatFlag,omitempty"`
}

type TickTickSubtask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	SortOrder int    `json:"sortOrder"`
}

type TickTickProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *TickTickClient) ListProjects(ctx context.Context) ([]TickTickProject, error) {
	data, status, err := c.doRequest(ctx, "GET", "/project", nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "ticktick", StatusCode: status, Message: string(data)}
	}
	var projects []TickTickProject
	return projects, json.Unmarshal(data, &projects)
}

func (c *TickTickClient) ListTasks(ctx context.Context, projectID string) ([]TickTickTask, error) {
	path := fmt.Sprintf("/project/%s/data", projectID)
	data, status, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "ticktick", StatusCode: status, Message: string(data)}
	}
	var result struct {
		Tasks []TickTickTask `json:"tasks"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Tasks, nil
}

func (c *TickTickClient) CreateTask(ctx context.Context, task *TickTickTask) (*TickTickTask, error) {
	data, status, err := c.doRequest(ctx, "POST", "/task", task)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "ticktick", StatusCode: status, Message: string(data)}
	}
	var created TickTickTask
	return &created, json.Unmarshal(data, &created)
}

func (c *TickTickClient) UpdateTask(ctx context.Context, task *TickTickTask) (*TickTickTask, error) {
	path := fmt.Sprintf("/task/%s", task.ID)
	data, status, err := c.doRequest(ctx, "POST", path, task)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "ticktick", StatusCode: status, Message: string(data)}
	}
	var updated TickTickTask
	return &updated, json.Unmarshal(data, &updated)
}

func (c *TickTickClient) DeleteTask(ctx context.Context, projectID, taskID string) error {
	path := fmt.Sprintf("/project/%s/task/%s", projectID, taskID)
	_, status, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return &salerr.APIError{Service: "ticktick", StatusCode: status, Message: "delete failed"}
	}
	return nil
}

// TickTickToCalendarItem maps a TickTick API task to the unified model.
func TickTickToCalendarItem(task TickTickTask) model.CalendarItem {
	item := model.CalendarItem{
		UID:         task.ID,
		Title:       task.Title,
		Description: task.Content,
		ItemType:    model.ItemTypeTask,
		Tags:        task.Tags,
	}

	if task.Desc != "" && item.Description == "" {
		item.Description = task.Desc
	}

	// Priority: TickTick 0=none, 1=low, 3=medium, 5=high
	switch task.Priority {
	case 5:
		item.Priority = model.PriorityHigh
	case 3:
		item.Priority = model.PriorityMedium
	case 1:
		item.Priority = model.PriorityLow
	default:
		item.Priority = model.PriorityNone
	}

	// Status: 0=normal, 2=completed
	if task.Status == 2 {
		item.Status = model.StatusCompleted
	} else {
		item.Status = model.StatusPending
	}

	if task.StartDate != "" {
		if t, err := parseTickTickDate(task.StartDate); err == nil {
			item.StartTime = &t
		}
	}
	if task.DueDate != "" {
		if t, err := parseTickTickDate(task.DueDate); err == nil {
			item.DueDate = &t
		}
	}

	for _, sub := range task.Items {
		st := model.Subtask{
			Title: sub.Title,
		}
		if sub.Status == 2 {
			st.Status = model.StatusCompleted
		}
		item.Subtasks = append(item.Subtasks, st)
	}

	return item
}

// CalendarItemToTickTick maps the unified model back to TickTick API format.
func CalendarItemToTickTick(item model.CalendarItem, projectID string) TickTickTask {
	task := TickTickTask{
		ID:        item.UID,
		ProjectID: projectID,
		Title:     item.Title,
		Content:   item.Description,
		Tags:      item.Tags,
	}

	switch item.Priority {
	case model.PriorityHighest, model.PriorityHigh:
		task.Priority = 5
	case model.PriorityMedium:
		task.Priority = 3
	case model.PriorityLow, model.PriorityLowest:
		task.Priority = 1
	default:
		task.Priority = 0
	}

	if item.Status == model.StatusCompleted {
		task.Status = 2
	}

	if item.StartTime != nil {
		task.StartDate = item.StartTime.UTC().Format("2006-01-02T15:04:05.000+0000")
	}
	if item.DueDate != nil {
		task.DueDate = item.DueDate.UTC().Format("2006-01-02T15:04:05.000+0000")
	}

	for _, sub := range item.Subtasks {
		st := TickTickSubtask{Title: sub.Title}
		if sub.Status == model.StatusCompleted {
			st.Status = 2
		}
		task.Items = append(task.Items, st)
	}

	return task
}

func parseTickTickDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse TickTick date: %s", s)
}
