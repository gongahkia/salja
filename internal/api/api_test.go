package api

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"
"time"

"github.com/gongahkia/calendar-converter/internal/model"
)

func newTestToken() *Token {
return &Token{
AccessToken: "test-token",
TokenType:   "Bearer",
ExpiresAt:   time.Now().Add(time.Hour),
}
}

func TestTickTickListTasks(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Header.Get("Authorization") != "Bearer test-token" {
t.Error("missing auth header")
}
json.NewEncoder(w).Encode(map[string]interface{}{
"tasks": []TickTickTask{
{ID: "t1", Title: "Test Task", Priority: 5, Status: 0, Tags: []string{"work"}},
},
})
}))
defer server.Close()

client := NewTickTickClient(newTestToken())
client.httpClient = server.Client()
	client.baseURL = server.URL

// Override base URL by making the request manually
ctx := context.Background()
data, status, err := client.doRequest(ctx, "GET", "/project/p1/data", nil)
if err != nil {
t.Fatal(err)
}
if status != 200 {
t.Fatalf("expected 200, got %d", status)
}
var result struct {
Tasks []TickTickTask `json:"tasks"`
}
json.Unmarshal(data, &result)
if len(result.Tasks) != 1 || result.Tasks[0].Title != "Test Task" {
t.Error("unexpected task data")
}
}

func TestTickTickMapper(t *testing.T) {
task := TickTickTask{
ID:       "abc123",
Title:    "Buy groceries",
Content:  "Milk, eggs, bread",
Priority: 5,
Status:   2,
Tags:     []string{"shopping"},
DueDate:  "2024-03-15T10:00:00.000+0000",
Items: []TickTickSubtask{
{Title: "Get milk", Status: 0},
{Title: "Get eggs", Status: 2},
},
}

item := TickTickToCalendarItem(task)
if item.Title != "Buy groceries" {
t.Errorf("title: got %q", item.Title)
}
if item.Priority != model.PriorityHighest {
t.Errorf("priority: got %d", item.Priority)
}
if item.Status != model.StatusCompleted {
t.Errorf("status: got %v", item.Status)
}
if len(item.Subtasks) != 2 {
t.Errorf("subtasks: got %d", len(item.Subtasks))
}
if item.DueDate == nil {
t.Error("due date should be set")
}

// Round-trip
back := CalendarItemToTickTick(item, "proj1")
if back.Title != "Buy groceries" {
t.Errorf("roundtrip title: got %q", back.Title)
}
if back.Priority != 5 {
t.Errorf("roundtrip priority: got %d", back.Priority)
}
}

func TestTodoistMapper(t *testing.T) {
due := &TodoistDue{Date: "2024-06-15", IsRecurring: false}
task := TodoistTask{
ID:       "td1",
Content:  "Review PR",
Priority: 4,
Labels:   []string{"dev"},
Due:      due,
}

item := TodoistToCalendarItem(task)
if item.Priority != model.PriorityHighest {
t.Errorf("priority: got %d", item.Priority)
}
if item.DueDate == nil {
t.Error("due date should be set")
}

back := CalendarItemToTodoist(item, "proj1")
if back.Priority != 4 {
t.Errorf("roundtrip priority: got %d", back.Priority)
}
}

func TestGCalMapper(t *testing.T) {
start := "2024-03-15T10:00:00Z"
end := "2024-03-15T11:00:00Z"
event := GCalEvent{
ID:       "gcal1",
Summary:  "Team Standup",
Location: "Room 42",
Start:    &GCalDateTime{DateTime: start},
End:      &GCalDateTime{DateTime: end},
Attendees: []GCalAttendee{{Email: "alice@example.com"}},
}

item := GCalToCalendarItem(event)
if item.Title != "Team Standup" {
t.Errorf("title: got %q", item.Title)
}
if item.StartTime == nil {
t.Error("start time should be set")
}
if len(item.Tags) != 1 {
t.Errorf("expected 1 attendee tag, got %d", len(item.Tags))
}

back := CalendarItemToGCal(item)
if back.Summary != "Team Standup" {
t.Errorf("roundtrip summary: got %q", back.Summary)
}
}

func TestMSGraphMapper(t *testing.T) {
event := MSGraphEvent{
ID:      "msg1",
Subject: "Budget Review",
Start:   &MSGraphDateTime{DateTime: "2024-03-15T14:00:00.0000000", TimeZone: "UTC"},
End:     &MSGraphDateTime{DateTime: "2024-03-15T15:00:00.0000000", TimeZone: "UTC"},
Body:    &MSGraphBody{ContentType: "text", Content: "Q1 numbers"},
Location: &MSGraphLocation{DisplayName: "Board Room"},
}

item := MSGraphToCalendarItem(event)
if item.Title != "Budget Review" {
t.Errorf("title: got %q", item.Title)
}
if item.Location != "Board Room" {
t.Errorf("location: got %q", item.Location)
}

back := CalendarItemToMSGraph(item)
if back.Subject != "Budget Review" {
t.Errorf("roundtrip: got %q", back.Subject)
}
}

func TestNotionMapper(t *testing.T) {
pm := DefaultNotionPropertyMap()
page := NotionPage{
ID: "n1",
Properties: map[string]NotionProperty{
"Name": {
Type:  "title",
Title: []NotionRichText{{PlainText: "Design mockups"}},
},
"Due Date": {
Type: "date",
Date: &NotionDate{Start: "2024-04-01"},
},
"Status": {
Type:   "status",
Status: &NotionSelect{Name: "In progress"},
},
"Priority": {
Type:   "select",
Select: &NotionSelect{Name: "High"},
},
"Tags": {
Type:        "multi_select",
MultiSelect: []NotionSelect{{Name: "design"}, {Name: "ui"}},
},
},
}

item := NotionToCalendarItem(page, pm)
if item.Title != "Design mockups" {
t.Errorf("title: got %q", item.Title)
}
if item.Status != model.StatusInProgress {
t.Errorf("status: got %v", item.Status)
}
if item.Priority != model.PriorityHigh {
t.Errorf("priority: got %d", item.Priority)
}
if len(item.Tags) != 2 {
t.Errorf("tags: got %d", len(item.Tags))
}

back := CalendarItemToNotion(item, pm)
if len(back.Properties) == 0 {
t.Error("properties should not be empty")
}
}

func TestTokenStoreRoundTrip(t *testing.T) {
dir := t.TempDir()
store := &TokenStore{Path: dir + "/tokens.json"}

token := &Token{
AccessToken:  "abc",
RefreshToken: "xyz",
TokenType:    "Bearer",
ExpiresAt:    time.Now().Add(time.Hour),
}

if err := store.Set("test-service", token); err != nil {
t.Fatal(err)
}

loaded, err := store.Get("test-service")
if err != nil {
t.Fatal(err)
}
if loaded.AccessToken != "abc" {
t.Errorf("expected 'abc', got %q", loaded.AccessToken)
}

if err := store.Delete("test-service"); err != nil {
t.Fatal(err)
}

_, err = store.Get("test-service")
if err == nil {
t.Error("expected error after delete")
}
}

func TestAPIError429(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(429)
w.Write([]byte(`{"message":"rate limited"}`))
}))
defer server.Close()

client := NewTickTickClient(newTestToken())
client.httpClient = server.Client()
	client.baseURL = server.URL

ctx := context.Background()
_, status, err := client.doRequest(ctx, "GET", "/project/p1/data", nil)
if err != nil {
t.Fatal(err)
}
if status != 429 {
t.Errorf("expected 429, got %d", status)
}
}
