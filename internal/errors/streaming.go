package errors

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	ical "github.com/emersion/go-ical"
)

// StreamingCSVParser reads CSV rows one at a time to handle large files.
type StreamingCSVParser struct {
	reader  *csv.Reader
	closer  io.Closer
	headers []string
	lineNum int
}

func NewStreamingCSVParser(r io.Reader) (*StreamingCSVParser, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1

	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var closer io.Closer
	if c, ok := r.(io.Closer); ok {
		closer = c
	}

	return &StreamingCSVParser{
		reader:  cr,
		closer:  closer,
		headers: headers,
		lineNum: 1,
	}, nil
}

func (p *StreamingCSVParser) Headers() []string {
	return p.headers
}

// Close releases any underlying resources.
func (p *StreamingCSVParser) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}

// Next returns the next row as a map, or io.EOF when done.
func (p *StreamingCSVParser) Next() (map[string]string, int, error) {
	record, err := p.reader.Read()
	if err != nil {
		return nil, p.lineNum, err
	}
	p.lineNum++

	row := make(map[string]string, len(p.headers))
	for i, h := range p.headers {
		if i < len(record) {
			row[h] = record[i]
		}
	}
	return row, p.lineNum, nil
}

// StreamingICSParser reads ICS components one at a time using the decoder.
type StreamingICSParser struct {
	decoder *ical.Decoder
	closer  io.Closer
}

func NewStreamingICSParser(r io.Reader) *StreamingICSParser {
	var closer io.Closer
	if c, ok := r.(io.Closer); ok {
		closer = c
	}
	return &StreamingICSParser{
		decoder: ical.NewDecoder(bufio.NewReader(r)),
		closer:  closer,
	}
}

// Next returns the next calendar component, or io.EOF when done.
func (p *StreamingICSParser) Next() (*ical.Calendar, error) {
	return p.decoder.Decode()
}

// Close releases any underlying resources.
func (p *StreamingICSParser) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
