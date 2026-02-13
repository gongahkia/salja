package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/model"
)

const gcalBaseURL = "https://www.googleapis.com/calendar/v3"

// GCalClient is a REST API client for Google Calendar.
type GCalClient struct {
	token      *Token
	httpClient *http.Client
}

func NewGCalClient(token *Token) *GCalClient {
	return NewGCalClientWithTimeout(token, 30*time.Second)
}

func NewGCalClientWithTimeout(token *Token, timeout time.Duration) *GCalClient {
	return &GCalClient{
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *GCalClient) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
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
		defer func() { _ = resp.Body.Close() }()

		respBody, err = io.ReadAll(resp.Body)
		statusCode = resp.StatusCode

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return &salerr.APIError{Service: "google-calendar", StatusCode: resp.StatusCode, Message: string(respBody)}
		}

		return err
	})

	return respBody, statusCode, err
}

// GCalEvent represents a Google Calendar event.
type GCalEvent struct {
	ID             string             `json:"id,omitempty"`
	Summary        string             `json:"summary"`
	Description    string             `json:"description,omitempty"`
	Location       string             `json:"location,omitempty"`
	Start          *GCalDateTime      `json:"start"`
	End            *GCalDateTime      `json:"end"`
	Recurrence     []string           `json:"recurrence,omitempty"`
	Attendees      []GCalAttendee     `json:"attendees,omitempty"`
	ConferenceData *GCalConference    `json:"conferenceData,omitempty"`
	Status         string             `json:"status,omitempty"`
	ExtendedProps  *GCalExtendedProps `json:"extendedProperties,omitempty"`
}

type GCalDateTime struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

type GCalAttendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
}

type GCalConference struct {
	EntryPoints []GCalEntryPoint `json:"entryPoints,omitempty"`
}

type GCalEntryPoint struct {
	EntryPointType string `json:"entryPointType"`
	URI            string `json:"uri"`
}

type GCalExtendedProps struct {
	Private map[string]string `json:"private,omitempty"`
	Shared  map[string]string `json:"shared,omitempty"`
}

type GCalCalendarList struct {
	Items []GCalCalendar `json:"items"`
}

type GCalCalendar struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Primary bool   `json:"primary"`
}

type GCalEventList struct {
	Items         []GCalEvent `json:"items"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

func (c *GCalClient) ListCalendars(ctx context.Context) ([]GCalCalendar, error) {
	data, status, err := c.doRequest(ctx, "GET", gcalBaseURL+"/users/me/calendarList", nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "Google Calendar", StatusCode: status, Message: string(data)}
	}
	var list GCalCalendarList
	return list.Items, json.Unmarshal(data, &list)
}

func (c *GCalClient) ListEvents(ctx context.Context, calendarID string, timeMin, timeMax time.Time) ([]GCalEvent, error) {
	url := fmt.Sprintf("%s/calendars/%s/events?timeMin=%s&timeMax=%s&singleEvents=false&maxResults=2500",
		gcalBaseURL, calendarID,
		timeMin.Format(time.RFC3339),
		timeMax.Format(time.RFC3339),
	)

	var allEvents []GCalEvent
	for url != "" {
		data, status, err := c.doRequest(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, &salerr.APIError{Service: "Google Calendar", StatusCode: status, Message: string(data)}
		}
		var list GCalEventList
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		allEvents = append(allEvents, list.Items...)
		if list.NextPageToken != "" {
			url = fmt.Sprintf("%s/calendars/%s/events?pageToken=%s", gcalBaseURL, calendarID, list.NextPageToken)
		} else {
			url = ""
		}
	}
	return allEvents, nil
}

func (c *GCalClient) InsertEvent(ctx context.Context, calendarID string, event *GCalEvent) (*GCalEvent, error) {
	url := fmt.Sprintf("%s/calendars/%s/events", gcalBaseURL, calendarID)
	data, status, err := c.doRequest(ctx, "POST", url, event)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "Google Calendar", StatusCode: status, Message: string(data)}
	}
	var created GCalEvent
	return &created, json.Unmarshal(data, &created)
}

func (c *GCalClient) UpdateEvent(ctx context.Context, calendarID string, event *GCalEvent) (*GCalEvent, error) {
	url := fmt.Sprintf("%s/calendars/%s/events/%s", gcalBaseURL, calendarID, event.ID)
	data, status, err := c.doRequest(ctx, "PUT", url, event)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, &salerr.APIError{Service: "Google Calendar", StatusCode: status, Message: string(data)}
	}
	var updated GCalEvent
	return &updated, json.Unmarshal(data, &updated)
}

func (c *GCalClient) DeleteEvent(ctx context.Context, calendarID, eventID string) error {
	url := fmt.Sprintf("%s/calendars/%s/events/%s", gcalBaseURL, calendarID, eventID)
	_, status, err := c.doRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	if status != 204 && status != 200 {
		return &salerr.APIError{Service: "Google Calendar", StatusCode: status, Message: "delete failed"}
	}
	return nil
}

// GCalToCalendarItem maps a Google Calendar event to the unified model.
func GCalToCalendarItem(event GCalEvent) model.CalendarItem {
	item := model.CalendarItem{
		UID:         event.ID,
		Title:       event.Summary,
		Description: event.Description,
		Location:    event.Location,
		ItemType:    model.ItemTypeEvent,
		Status:      model.StatusPending,
	}

	if event.Status == "cancelled" {
		item.Status = model.StatusCancelled
	}

	if event.Start != nil {
		if t := parseGCalTime(event.Start); t != nil {
			item.StartTime = t
			if event.Start.Date != "" {
				item.IsAllDay = true
			}
		}
	}
	if event.End != nil {
		if t := parseGCalTime(event.End); t != nil {
			item.EndTime = t
		}
	}

	if len(event.Recurrence) > 0 {
		for _, rule := range event.Recurrence {
			if len(rule) > 6 && rule[:6] == "RRULE:" {
				if rec, err := parseRRuleString(rule[6:]); err == nil {
					item.Recurrence = rec
				}
			}
		}
	}

	// Map conference link to description
	if event.ConferenceData != nil {
		for _, ep := range event.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				item.Description += "\nVideo: " + ep.URI
				break
			}
		}
	}

	// Map attendees to tags for interop
	for _, a := range event.Attendees {
		item.Tags = append(item.Tags, "attendee:"+a.Email)
	}

	return item
}

// CalendarItemToGCal maps the unified model to a Google Calendar event.
func CalendarItemToGCal(item model.CalendarItem) GCalEvent {
	event := GCalEvent{
		ID:          item.UID,
		Summary:     item.Title,
		Description: item.Description,
		Location:    item.Location,
	}

	if item.StartTime != nil {
		event.Start = &GCalDateTime{}
		if item.IsAllDay {
			event.Start.Date = item.StartTime.Format("2006-01-02")
		} else {
			event.Start.DateTime = item.StartTime.Format(time.RFC3339)
		}
	}
	if item.EndTime != nil {
		event.End = &GCalDateTime{}
		if item.IsAllDay {
			event.End.Date = item.EndTime.Format("2006-01-02")
		} else {
			event.End.DateTime = item.EndTime.Format(time.RFC3339)
		}
	}

	return event
}

func parseGCalTime(dt *GCalDateTime) *time.Time {
	if dt.DateTime != "" {
		if t, err := time.Parse(time.RFC3339, dt.DateTime); err == nil {
			return &t
		}
	}
	if dt.Date != "" {
		if t, err := time.Parse("2006-01-02", dt.Date); err == nil {
			return &t
		}
	}
	return nil
}

func parseRRuleString(value string) (*model.Recurrence, error) {
	rec := &model.Recurrence{Interval: 1}
	parts := strings.Split(value, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "FREQ":
			rec.Freq = model.FreqType(kv[1])
		case "INTERVAL":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				rec.Interval = v
			}
		case "COUNT":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				rec.Count = &v
			}
		case "UNTIL":
			if t, err := time.Parse("20060102T150405Z", kv[1]); err == nil {
				rec.Until = &t
			}
		case "BYDAY":
			for _, d := range strings.Split(kv[1], ",") {
				rec.ByDay = append(rec.ByDay, model.Weekday(strings.TrimSpace(d)))
			}
		}
	}
	return rec, nil
}
