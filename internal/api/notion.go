package api

import (
"bytes"
"context"
"encoding/json"
"io"
"net/http"
"time"

salerr "github.com/gongahkia/salja/internal/errors"
"github.com/gongahkia/salja/internal/model"
)

const notionBaseURL = "https://api.notion.com/v1"
const notionVersion = "2022-06-28"

// NotionClient is an API client for Notion.
type NotionClient struct {
token      string
httpClient *http.Client
}

func NewNotionClient(bearerToken string) *NotionClient {
return &NotionClient{
token:      bearerToken,
httpClient: &http.Client{Timeout: 30 * time.Second},
}
}

func NewNotionClientWithTimeout(bearerToken string, timeout time.Duration) *NotionClient {
return &NotionClient{
token:      bearerToken,
httpClient: &http.Client{Timeout: timeout},
}
}

func (c *NotionClient) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
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
req.Header.Set("Authorization", "Bearer "+c.token)
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Notion-Version", notionVersion)

resp, err := c.httpClient.Do(req)
if err != nil {
return nil, 0, err
}
defer resp.Body.Close()

respBody, err := io.ReadAll(resp.Body)
return respBody, resp.StatusCode, err
}

// NotionPage represents a page in a Notion database.
type NotionPage struct {
ID         string                    `json:"id,omitempty"`
Properties map[string]NotionProperty `json:"properties"`
}

// NotionProperty is a generic Notion property value.
type NotionProperty struct {
Type     string             `json:"type"`
Title    []NotionRichText   `json:"title,omitempty"`
RichText []NotionRichText   `json:"rich_text,omitempty"`
Date     *NotionDate        `json:"date,omitempty"`
Select   *NotionSelect      `json:"select,omitempty"`
Status   *NotionSelect      `json:"status,omitempty"`
Checkbox *bool              `json:"checkbox,omitempty"`
Number   *float64           `json:"number,omitempty"`
MultiSelect []NotionSelect  `json:"multi_select,omitempty"`
}

type NotionRichText struct {
PlainText string `json:"plain_text,omitempty"`
Text      *struct {
Content string `json:"content"`
} `json:"text,omitempty"`
}

type NotionDate struct {
Start string `json:"start"`
End   string `json:"end,omitempty"`
}

type NotionSelect struct {
Name string `json:"name"`
}

type NotionQueryResult struct {
Results    []NotionPage `json:"results"`
HasMore    bool         `json:"has_more"`
NextCursor string      `json:"next_cursor,omitempty"`
}

// NotionPropertyMap configures which Notion properties map to CalendarItem fields.
type NotionPropertyMap struct {
Title       string `toml:"title" json:"title"`
Description string `toml:"description" json:"description"`
DueDate     string `toml:"due_date" json:"due_date"`
Status      string `toml:"status" json:"status"`
Priority    string `toml:"priority" json:"priority"`
Tags        string `toml:"tags" json:"tags"`
}

func DefaultNotionPropertyMap() NotionPropertyMap {
return NotionPropertyMap{
Title:       "Name",
Description: "Description",
DueDate:     "Due Date",
Status:      "Status",
Priority:    "Priority",
Tags:        "Tags",
}
}

func (c *NotionClient) QueryDatabase(ctx context.Context, databaseID string, startCursor string) (*NotionQueryResult, error) {
body := map[string]interface{}{}
if startCursor != "" {
body["start_cursor"] = startCursor
}
body["page_size"] = 100

data, status, err := c.doRequest(ctx, "POST", notionBaseURL+"/databases/"+databaseID+"/query", body)
if err != nil {
return nil, err
}
if status != 200 {
return nil, &salerr.APIError{Service: "notion", StatusCode: status, Message: string(data)}
}
var result NotionQueryResult
return &result, json.Unmarshal(data, &result)
}

func (c *NotionClient) CreatePage(ctx context.Context, databaseID string, page *NotionPage) (*NotionPage, error) {
body := map[string]interface{}{
"parent":     map[string]string{"database_id": databaseID},
"properties": page.Properties,
}

data, status, err := c.doRequest(ctx, "POST", notionBaseURL+"/pages", body)
if err != nil {
return nil, err
}
if status != 200 {
return nil, &salerr.APIError{Service: "notion", StatusCode: status, Message: string(data)}
}
var created NotionPage
return &created, json.Unmarshal(data, &created)
}

func (c *NotionClient) UpdatePage(ctx context.Context, pageID string, properties map[string]NotionProperty) error {
body := map[string]interface{}{
"properties": properties,
}

_, status, err := c.doRequest(ctx, "PATCH", notionBaseURL+"/pages/"+pageID, body)
if err != nil {
return err
}
if status != 200 {
return &salerr.APIError{Service: "notion", StatusCode: status, Message: "update failed"}
}
return nil
}

// NotionToCalendarItem maps a Notion page to the unified model using property mapping.
func NotionToCalendarItem(page NotionPage, pm NotionPropertyMap) model.CalendarItem {
item := model.CalendarItem{
UID:      page.ID,
ItemType: model.ItemTypeTask,
Status:   model.StatusPending,
}

if prop, ok := page.Properties[pm.Title]; ok && len(prop.Title) > 0 {
item.Title = prop.Title[0].PlainText
}

if prop, ok := page.Properties[pm.Description]; ok && len(prop.RichText) > 0 {
item.Description = prop.RichText[0].PlainText
}

if prop, ok := page.Properties[pm.DueDate]; ok && prop.Date != nil {
if t, err := time.Parse("2006-01-02", prop.Date.Start); err == nil {
item.DueDate = &t
} else if t, err := time.Parse(time.RFC3339, prop.Date.Start); err == nil {
item.DueDate = &t
}
}

if prop, ok := page.Properties[pm.Status]; ok && prop.Status != nil {
switch prop.Status.Name {
case "Done", "Completed", "Complete":
item.Status = model.StatusCompleted
case "In Progress", "In progress":
item.Status = model.StatusInProgress
}
}

if prop, ok := page.Properties[pm.Priority]; ok && prop.Select != nil {
switch prop.Select.Name {
case "Urgent", "Critical":
item.Priority = model.PriorityHighest
case "High":
item.Priority = model.PriorityHigh
case "Medium":
item.Priority = model.PriorityMedium
case "Low":
item.Priority = model.PriorityLow
}
}

if prop, ok := page.Properties[pm.Tags]; ok {
for _, s := range prop.MultiSelect {
item.Tags = append(item.Tags, s.Name)
}
}

return item
}

// CalendarItemToNotion maps the unified model to a Notion page.
func CalendarItemToNotion(item model.CalendarItem, pm NotionPropertyMap) NotionPage {
props := make(map[string]NotionProperty)

props[pm.Title] = NotionProperty{
Type: "title",
Title: []NotionRichText{{Text: &struct {
Content string `json:"content"`
}{Content: item.Title}}},
}

if item.Description != "" {
props[pm.Description] = NotionProperty{
Type: "rich_text",
RichText: []NotionRichText{{Text: &struct {
Content string `json:"content"`
}{Content: item.Description}}},
}
}

if item.DueDate != nil {
props[pm.DueDate] = NotionProperty{
Type: "date",
Date: &NotionDate{Start: item.DueDate.Format("2006-01-02")},
}
}

statusName := "Not started"
switch item.Status {
case model.StatusCompleted:
statusName = "Done"
case model.StatusInProgress:
statusName = "In progress"
}
props[pm.Status] = NotionProperty{
Type:   "status",
Status: &NotionSelect{Name: statusName},
}

if item.Priority > 0 {
pName := "Low"
switch item.Priority {
case model.PriorityHighest:
pName = "Urgent"
case model.PriorityHigh:
pName = "High"
case model.PriorityMedium:
pName = "Medium"
}
props[pm.Priority] = NotionProperty{
Type:   "select",
Select: &NotionSelect{Name: pName},
}
}

if len(item.Tags) > 0 {
var ms []NotionSelect
for _, t := range item.Tags {
ms = append(ms, NotionSelect{Name: t})
}
props[pm.Tags] = NotionProperty{
Type:        "multi_select",
MultiSelect: ms,
}
}

return NotionPage{
ID:         item.UID,
Properties: props,
}
}
