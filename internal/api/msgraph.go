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

const graphBaseURL = "https://graph.microsoft.com/v1.0"

// MSGraphClient is a REST API client for Microsoft Graph (Outlook Calendar).
type MSGraphClient struct {
	token      *Token
	httpClient *http.Client
}

func NewMSGraphClient(token *Token) *MSGraphClient {
	return NewMSGraphClientWithTimeout(token, 30*time.Second)
}

func NewMSGraphClientWithTimeout(token *Token, timeout time.Duration) *MSGraphClient {
	return &MSGraphClient{
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *MSGraphClient) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
	var respBody []byte
	var statusCode int

	err := salerr.Retry(salerr.DefaultRetryConfig(), func() error {
		var reqBody io.Reader
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return err
			}
			reqBody = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBody, err = io.ReadAll(resp.Body)
		statusCode = resp.StatusCode

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return &salerr.APIError{Service: "microsoft-graph", StatusCode: resp.StatusCode, Message: string(respBody)}
		}

		return err
	})

	return respBody, statusCode, err
}

// MSGraphEvent represents an Outlook calendar event.
type MSGraphEvent struct {
	ID          string             `json:"id,omitempty"`
	Subject     string             `json:"subject"`
	Body        *MSGraphBody       `json:"body,omitempty"`
	Start       *MSGraphDateTime   `json:"start"`
	End         *MSGraphDateTime   `json:"end"`
	Location    *MSGraphLocation   `json:"location,omitempty"`
	IsAllDay    bool               `json:"isAllDay"`
	Recurrence  *MSGraphRecurrence `json:"recurrence,omitempty"`
	IsCancelled bool               `json:"isCancelled"`
}

type MSGraphBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type MSGraphDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type MSGraphLocation struct {
	DisplayName string `json:"displayName"`
}

type MSGraphRecurrence struct {
	Pattern *MSGraphRecurrencePattern `json:"pattern,omitempty"`
	Range   *MSGraphRecurrenceRange   `json:"range,omitempty"`
}

type MSGraphRecurrencePattern struct {
	Type       string   `json:"type"`
	Interval   int      `json:"interval"`
	DaysOfWeek []string `json:"daysOfWeek,omitempty"`
}

type MSGraphRecurrenceRange struct {
	Type      string `json:"type"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate,omitempty"`
}

type MSGraphEventList struct {
	Value    []MSGraphEvent `json:"value"`
	NextLink string         `json:"@odata.nextLink,omitempty"`
}

func (c *MSGraphClient) ListEvents(ctx context.Context, startTime, endTime time.Time) ([]MSGraphEvent, error) {
	url := fmt.Sprintf("%s/me/calendarview?startdatetime=%s&enddatetime=%s&$top=100",
		graphBaseURL,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	var allEvents []MSGraphEvent
	for url != "" {
		data, status, err := c.doRequest(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, &salerr.APIError{Service: "Microsoft Graph", StatusCode: status, Message: string(data)}
		}
		var list MSGraphEventList
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		allEvents = append(allEvents, list.Value...)
		url = list.NextLink
	}
	return allEvents, nil
}

func (c *MSGraphClient) CreateEvent(ctx context.Context, event *MSGraphEvent) (*MSGraphEvent, error) {
	data, status, err := c.doRequest(ctx, "POST", graphBaseURL+"/me/events", event)
	if err != nil {
		return nil, err
	}
	if status != 201 {
		return nil, &salerr.APIError{Service: "Microsoft Graph", StatusCode: status, Message: string(data)}
	}
	var created MSGraphEvent
	return &created, json.Unmarshal(data, &created)
}

func (c *MSGraphClient) UpdateEvent(ctx context.Context, event *MSGraphEvent) (*MSGraphEvent, error) {
	url := fmt.Sprintf("%s/me/events/%s", graphBaseURL, event.ID)
	data, status, err := c.doRequest(ctx, "PATCH", url, event)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "Microsoft Graph", StatusCode: status, Message: string(data)}
	}
	var updated MSGraphEvent
	return &updated, json.Unmarshal(data, &updated)
}

// MSGraphToCalendarItem maps an Outlook event to the unified model.
func MSGraphToCalendarItem(event MSGraphEvent) model.CalendarItem {
	item := model.CalendarItem{
		UID:      event.ID,
		Title:    event.Subject,
		ItemType: model.ItemTypeEvent,
		IsAllDay: event.IsAllDay,
		Status:   model.StatusPending,
	}

	if event.IsCancelled {
		item.Status = model.StatusCancelled
	}

	if event.Body != nil {
		item.Description = event.Body.Content
	}
	if event.Location != nil {
		item.Location = event.Location.DisplayName
	}

	if event.Start != nil {
		if t, err := time.Parse("2006-01-02T15:04:05.0000000", event.Start.DateTime); err == nil {
			if event.Start.TimeZone != "" && event.Start.TimeZone != "UTC" {
				if loc, locErr := time.LoadLocation(event.Start.TimeZone); locErr == nil {
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
				}
				item.Timezone = event.Start.TimeZone
			}
			item.StartTime = &t
		}
	}
	if event.End != nil {
		if t, err := time.Parse("2006-01-02T15:04:05.0000000", event.End.DateTime); err == nil {
			if event.End.TimeZone != "" && event.End.TimeZone != "UTC" {
				if loc, locErr := time.LoadLocation(event.End.TimeZone); locErr == nil {
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
				}
			}
			item.EndTime = &t
		}
	}

	return item
}

// CalendarItemToMSGraph maps the unified model to an Outlook event.
func CalendarItemToMSGraph(item model.CalendarItem) MSGraphEvent {
	event := MSGraphEvent{
		ID:       item.UID,
		Subject:  item.Title,
		IsAllDay: item.IsAllDay,
	}

	if item.Description != "" {
		event.Body = &MSGraphBody{ContentType: "text", Content: item.Description}
	}
	if item.Location != "" {
		event.Location = &MSGraphLocation{DisplayName: item.Location}
	}

	if item.StartTime != nil {
		tz := "UTC"
		if item.Timezone != "" {
			tz = item.Timezone
		}
		event.Start = &MSGraphDateTime{
			DateTime: item.StartTime.Format("2006-01-02T15:04:05.0000000"),
			TimeZone: tz,
		}
	}
	if item.EndTime != nil {
		tz := "UTC"
		if item.Timezone != "" {
			tz = item.Timezone
		}
		event.End = &MSGraphDateTime{
			DateTime: item.EndTime.Format("2006-01-02T15:04:05.0000000"),
			TimeZone: tz,
		}
	}

	return event
}
