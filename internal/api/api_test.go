package api

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"net/url"
"strings"
"sync/atomic"
"testing"
"time"

"github.com/gongahkia/salja/internal/model"
)

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// redirectClient returns an *http.Client whose transport rewrites every
// outgoing request so that it hits the given httptest server instead of
// the real host embedded in the URL.
func redirectClient(ts *httptest.Server) *http.Client {
	tsURL, _ := url.Parse(ts.URL)
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = tsURL.Scheme
			req.URL.Host = tsURL.Host
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

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
if item.Priority != model.PriorityHigh {
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

// ---------------------------------------------------------------------------
// Todoist httptest tests
// ---------------------------------------------------------------------------

func TestTodoistGetTasks(t *testing.T) {
	tasks := []TodoistTask{
		{ID: "1", Content: "Buy milk", Priority: 3, Labels: []string{"errands"}, Due: &TodoistDue{Date: "2024-08-01"}},
		{ID: "2", Content: "Write tests", Priority: 4, IsCompleted: true},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/tasks") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong auth header")
		}
		json.NewEncoder(w).Encode(tasks)
	}))
	defer ts.Close()

	client := NewTodoistClient(newTestToken())
	client.httpClient = redirectClient(ts)

	got, err := client.GetTasks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}
	if got[0].Content != "Buy milk" {
		t.Errorf("task[0].Content = %q", got[0].Content)
	}
	if got[1].IsCompleted != true {
		t.Error("task[1] should be completed")
	}

	item := TodoistToCalendarItem(got[0])
	if item.Priority != model.PriorityHigh {
		t.Errorf("mapped priority: got %d, want %d", item.Priority, model.PriorityHigh)
	}
	if item.DueDate == nil {
		t.Error("due date should be set")
	}
}

func TestTodoistCreateTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var incoming TodoistTask
		json.NewDecoder(r.Body).Decode(&incoming)
		incoming.ID = "new-123"
		json.NewEncoder(w).Encode(incoming)
	}))
	defer ts.Close()

	client := NewTodoistClient(newTestToken())
	client.httpClient = redirectClient(ts)

	created, err := client.CreateTask(context.Background(), &TodoistTask{
		Content:  "Deploy v2",
		Priority: 2,
		Labels:   []string{"ops"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "new-123" {
		t.Errorf("expected ID 'new-123', got %q", created.ID)
	}
	if created.Content != "Deploy v2" {
		t.Errorf("content: got %q", created.Content)
	}
}

func TestTodoistGetTasksNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer ts.Close()

	client := NewTodoistClient(newTestToken())
	client.httpClient = redirectClient(ts)

	_, err := client.GetTasks(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status 500: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Notion httptest tests
// ---------------------------------------------------------------------------

func TestNotionQueryDatabase(t *testing.T) {
	var callCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Notion-Version") != notionVersion {
			t.Errorf("missing Notion-Version header")
		}

		call := atomic.AddInt32(&callCount, 1)
		if call == 1 {
			json.NewEncoder(w).Encode(NotionQueryResult{
				Results: []NotionPage{
					{ID: "page-1", Properties: map[string]NotionProperty{
						"Name": {Type: "title", Title: []NotionRichText{{PlainText: "First"}}},
					}},
				},
				HasMore:    true,
				NextCursor: "cursor-abc",
			})
		} else {
			// Verify cursor was sent
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["start_cursor"] != "cursor-abc" {
				t.Errorf("expected start_cursor 'cursor-abc', got %v", body["start_cursor"])
			}
			json.NewEncoder(w).Encode(NotionQueryResult{
				Results: []NotionPage{
					{ID: "page-2", Properties: map[string]NotionProperty{
						"Name": {Type: "title", Title: []NotionRichText{{PlainText: "Second"}}},
					}},
				},
				HasMore: false,
			})
		}
	}))
	defer ts.Close()

	client := NewNotionClient("test-notion-token")
	client.httpClient = redirectClient(ts)
	ctx := context.Background()

	// First page
	result1, err := client.QueryDatabase(ctx, "db-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result1.Results) != 1 || result1.Results[0].ID != "page-1" {
		t.Errorf("page 1 unexpected: %+v", result1)
	}
	if !result1.HasMore {
		t.Error("expected HasMore=true on first call")
	}

	// Second page using cursor
	result2, err := client.QueryDatabase(ctx, "db-1", result1.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(result2.Results) != 1 || result2.Results[0].ID != "page-2" {
		t.Errorf("page 2 unexpected: %+v", result2)
	}
	if result2.HasMore {
		t.Error("expected HasMore=false on second call")
	}

	// Verify mapping of first result
	pm := DefaultNotionPropertyMap()
	item := NotionToCalendarItem(result1.Results[0], pm)
	if item.Title != "First" {
		t.Errorf("mapped title: got %q", item.Title)
	}
}

func TestNotionCreatePage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/pages") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(NotionPage{
			ID: "created-page-1",
			Properties: map[string]NotionProperty{
				"Name": {Type: "title", Title: []NotionRichText{{PlainText: "New Task"}}},
			},
		})
	}))
	defer ts.Close()

	client := NewNotionClient("test-notion-token")
	client.httpClient = redirectClient(ts)

	pm := DefaultNotionPropertyMap()
	dueDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	page := CalendarItemToNotion(model.CalendarItem{
		Title:    "New Task",
		DueDate:  &dueDate,
		Priority: model.PriorityMedium,
		Status:   model.StatusPending,
	}, pm)

	created, err := client.CreatePage(context.Background(), "db-1", &page)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "created-page-1" {
		t.Errorf("expected ID 'created-page-1', got %q", created.ID)
	}
	item := NotionToCalendarItem(*created, pm)
	if item.Title != "New Task" {
		t.Errorf("mapped title: got %q", item.Title)
	}
}

func TestNotionQueryDatabase429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"message":"rate limited","code":"rate_limited"}`))
	}))
	defer ts.Close()

	client := NewNotionClient("test-notion-token")
	client.httpClient = redirectClient(ts)

	_, err := client.QueryDatabase(context.Background(), "db-1", "")
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error should mention 429: %v", err)
	}
	if !strings.Contains(err.Error(), "rate") {
		t.Errorf("error should contain rate limit message: %v", err)
	}
}
