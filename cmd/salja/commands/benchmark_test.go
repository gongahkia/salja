package commands_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/ics"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
	"github.com/gongahkia/salja/internal/writers"
)

const benchItemCount = 10000

// generateGCalCSV creates a Google Calendar CSV with n events.
func generateGCalCSV(n int) []byte {
	var buf bytes.Buffer
	buf.WriteString("Subject,Start Date,Start Time,End Date,End Time,All Day Event,Description,Location\n")
	base := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		start := base.Add(time.Duration(i) * time.Hour)
		end := start.Add(time.Hour)
		fmt.Fprintf(&buf, "Event %d,%s,%s,%s,%s,False,Description for event %d,Room %d\n",
			i,
			start.Format("01/02/2006"),
			start.Format("3:04 PM"),
			end.Format("01/02/2006"),
			end.Format("3:04 PM"),
			i, i%10,
		)
	}
	return buf.Bytes()
}

// generateICSData creates ICS content with n events.
func generateICSData(n int) []byte {
	var buf bytes.Buffer
	buf.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//benchmark//test//EN\r\n")
	base := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		start := base.Add(time.Duration(i) * time.Hour)
		end := start.Add(time.Hour)
		fmt.Fprintf(&buf, "BEGIN:VEVENT\r\nUID:bench-%d@test\r\nSUMMARY:Event %d\r\nDTSTART:%s\r\nDTEND:%s\r\nLOCATION:Room %d\r\nDESCRIPTION:Description %d\r\nEND:VEVENT\r\n",
			i, i,
			start.Format("20060102T150405Z"),
			end.Format("20060102T150405Z"),
			i%10, i,
		)
	}
	buf.WriteString("END:VCALENDAR\r\n")
	return buf.Bytes()
}

// generateTickTickCSV creates a TickTick CSV with n tasks.
func generateTickTickCSV(n int) []byte {
	var buf bytes.Buffer
	buf.WriteString("\"Folder Name\",\"List Name\",\"Title\",\"Tags\",\"Content\",\"Is Check list\",\"Start Date\",\"Due Date\",\"Reminder\",\"Repeat\",\"Priority\",\"Status\",\"Created Time\",\"Completed Time\",\"Order\",\"Timezone\",\"Is All Day\"\n")
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	priorities := []string{"0", "1", "3", "5"}
	for i := 0; i < n; i++ {
		due := base.Add(time.Duration(i) * 24 * time.Hour)
		fmt.Fprintf(&buf, "\"\",\"Inbox\",\"Task %d\",\"tag%d\",\"Content %d\",\"N\",\"\",\"%s\",\"\",\"\",\"%s\",\"0\",\"%s\",\"\",\"%d\",\"UTC\",\"false\"\n",
			i, i%5, i,
			due.Format("2006-01-02T15:04:05+0000"),
			priorities[i%4],
			base.Format("2006-01-02T15:04:05+0000"),
			i,
		)
	}
	return buf.Bytes()
}

func BenchmarkGCalCSVParser(b *testing.B) {
	data := generateGCalCSV(benchItemCount)
	ctx := context.Background()
	parser := parsers.NewGoogleCalendarParser()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := parser.Parse(ctx, r, "bench.csv")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkICSParser(b *testing.B) {
	data := generateICSData(benchItemCount)
	ctx := context.Background()
	parser := ics.NewParser()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := parser.Parse(ctx, r, "bench.ics")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTickTickCSVParser(b *testing.B) {
	data := generateTickTickCSV(benchItemCount)
	ctx := context.Background()
	parser := parsers.NewTickTickParser()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := parser.Parse(ctx, r, "ticktick.csv")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGCalCSVWriter(b *testing.B) {
	collection := generateLargeCollection(benchItemCount, model.ItemTypeEvent)
	ctx := context.Background()
	writer := writers.NewGoogleCalendarWriter()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := writer.Write(ctx, collection, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkICSWriter(b *testing.B) {
	collection := generateLargeCollection(benchItemCount, model.ItemTypeEvent)
	ctx := context.Background()
	writer := ics.NewWriter()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := writer.Write(ctx, collection, &buf); err != nil {
			b.Fatal(err)
		}
	}
}

func generateLargeCollection(n int, itemType model.ItemType) *model.CalendarCollection {
	base := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	items := make([]model.CalendarItem, n)
	for i := range items {
		start := base.Add(time.Duration(i) * time.Hour)
		end := start.Add(time.Hour)
		items[i] = model.CalendarItem{
			UID:         fmt.Sprintf("bench-%d", i),
			Title:       fmt.Sprintf("Item %d", i),
			Description: strings.Repeat("x", 50),
			StartTime:   &start,
			EndTime:     &end,
			Location:    fmt.Sprintf("Room %d", i%10),
			ItemType:    itemType,
			Status:      model.StatusPending,
		}
	}
	return &model.CalendarCollection{Items: items, SourceApp: "benchmark"}
}
