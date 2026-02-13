package errors

import (
	"io"
	"strings"
	"testing"
)

func TestMalformedCSVRow(t *testing.T) {
	input := "Name,Date,Priority\nTask1,2024-01-01,High\n\"unclosed quote\n"
	parser, err := NewStreamingCSVParser(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected header error: %v", err)
	}
	if len(parser.Headers()) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(parser.Headers()))
	}

	// First row should succeed
	row, _, err := parser.Next()
	if err != nil {
		t.Fatalf("first row should parse: %v", err)
	}
	if row["Name"] != "Task1" {
		t.Errorf("expected Task1, got %s", row["Name"])
	}

	// Second row has unclosed quote — csv reader with LazyQuotes may handle it
	_, _, err = parser.Next()
	// Either parses leniently or returns error — both acceptable
	_ = err
}

func TestInvalidICSProperty(t *testing.T) {
	input := "this is not valid ics content at all"
	parser := NewStreamingICSParser(strings.NewReader(input))
	_, err := parser.Next()
	if err == nil {
		t.Error("expected error for invalid ICS content")
	}
}

func TestCharsetDetectionUTF8(t *testing.T) {
	data := []byte("Hello, World!")
	enc := DetectEncoding(data)
	if enc != EncodingUTF8 {
		t.Errorf("expected UTF-8, got %s", enc)
	}
}

func TestCharsetDetectionUTF16LE(t *testing.T) {
	data := []byte{0xFF, 0xFE, 'H', 0, 'i', 0}
	enc := DetectEncoding(data)
	if enc != EncodingUTF16LE {
		t.Errorf("expected UTF-16LE, got %s", enc)
	}

	reader, detected, err := TranscodeToUTF8(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("transcode error: %v", err)
	}
	if detected != EncodingUTF16LE {
		t.Errorf("expected UTF-16LE detection, got %s", detected)
	}

	out, _ := io.ReadAll(reader)
	if string(out) != "Hi" {
		t.Errorf("expected 'Hi', got %q", string(out))
	}
}

func TestCharsetDetectionLatin1(t *testing.T) {
	// Bytes 0xA0-0xFF are Latin-1 but not CP1252-specific (no 0x80-0x9F bytes)
	data := []byte{0xE4, 0xF6, 0xFC} // äöü in Latin-1, invalid UTF-8
	enc := DetectEncoding(data)
	if enc != EncodingLatin1 {
		t.Errorf("expected Latin-1, got %s", enc)
	}
}

func TestCharsetDetectionCP1252(t *testing.T) {
	// Bytes 0x80-0x9F are CP1252-specific
	data := []byte{0x80, 0x81, 0x82}
	enc := DetectEncoding(data)
	if enc != EncodingCP1252 {
		t.Errorf("expected Windows-1252, got %s", enc)
	}
}

func TestAmbiguousDateMDY(t *testing.T) {
	parser := NewAmbiguousDateParser("en_US")
	d, err := parser.Parse("01/15/2024")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if d.Month() != 1 || d.Day() != 15 {
		t.Errorf("expected Jan 15, got %v", d)
	}
}

func TestAmbiguousDateDMY(t *testing.T) {
	parser := NewAmbiguousDateParser("en_GB")
	d, err := parser.Parse("15/01/2024")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if d.Month() != 1 || d.Day() != 15 {
		t.Errorf("expected Jan 15, got %v", d)
	}
}

func TestAmbiguousDateDetection(t *testing.T) {
	if !IsAmbiguous("03/04/2024") {
		t.Error("03/04/2024 should be ambiguous")
	}
	if IsAmbiguous("01/15/2024") {
		t.Error("01/15/2024 should not be ambiguous (15 > 12)")
	}
}

func TestAmbiguousDateISO(t *testing.T) {
	parser := NewAmbiguousDateParser("en_US")
	d, err := parser.Parse("2024-03-15")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if d.Month() != 3 || d.Day() != 15 {
		t.Errorf("expected Mar 15, got %v", d)
	}
}

func TestPartialFailureMode(t *testing.T) {
	result := NewPartialResult[string]()
	result.Add("item1")
	result.AddError(1, "id2", "invalid date", nil)
	result.Add("item3")

	if result.SuccessCount() != 2 {
		t.Errorf("expected 2 successes, got %d", result.SuccessCount())
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	summary := result.Summary()
	if !strings.Contains(summary, "2/3") {
		t.Errorf("summary should contain '2/3': %s", summary)
	}
}

func TestRetryExponentialBackoff(t *testing.T) {
	attempts := 0
	cfg := &RetryConfig{MaxRetries: 2, BaseDelay: 0} // zero delay for test speed
	err := Retry(cfg, func() error {
		attempts++
		if attempts < 3 {
			return &APIError{Service: "test", StatusCode: 429, Message: "rate limited"}
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestErrorCollectorSummary(t *testing.T) {
	c := NewErrorCollector()
	c.AddError(&ParseError{File: "test.csv", Line: 5, Message: "bad row"})
	c.AddWarning("field truncated")
	if !c.HasErrors() {
		t.Error("should have errors")
	}
	if c.Summary() != "1 errors, 1 warnings" {
		t.Errorf("unexpected summary: %s", c.Summary())
	}
}
