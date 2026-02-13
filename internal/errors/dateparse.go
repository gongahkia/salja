package errors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DateFormat represents a date format convention.
type DateFormat int

const (
	DateFormatMDY DateFormat = iota // MM/DD/YYYY (US)
	DateFormatDMY                   // DD/MM/YYYY (EU)
	DateFormatYMD                   // YYYY-MM-DD (ISO)
)

// AmbiguousDateParser handles parsing of dates with ambiguous formats.
type AmbiguousDateParser struct {
	PreferredFormat DateFormat
	Locale          string
}

func NewAmbiguousDateParser(locale string) *AmbiguousDateParser {
	format := DateFormatMDY
	locale = strings.ToLower(locale)
	switch {
	case strings.HasPrefix(locale, "en_gb"), strings.HasPrefix(locale, "en-gb"),
		strings.HasPrefix(locale, "de"), strings.HasPrefix(locale, "fr"),
		strings.HasPrefix(locale, "es"), strings.HasPrefix(locale, "it"),
		strings.HasPrefix(locale, "pt"), strings.HasPrefix(locale, "nl"):
		format = DateFormatDMY
	case strings.HasPrefix(locale, "ja"), strings.HasPrefix(locale, "zh"),
		strings.HasPrefix(locale, "ko"):
		format = DateFormatYMD
	}
	return &AmbiguousDateParser{PreferredFormat: format, Locale: locale}
}

var (
	slashDateRE = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{2,4})$`)
	dashDateRE  = regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})$`)
	dotDateRE   = regexp.MustCompile(`^(\d{1,2})\.(\d{1,2})\.(\d{2,4})$`)
)

// Parse attempts to parse an ambiguous date string.
func (p *AmbiguousDateParser) Parse(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	// Try ISO 8601 first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// YYYY-MM-DD
	if m := dashDateRE.FindStringSubmatch(s); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC), nil
	}

	// Slash separated: ambiguous
	if m := slashDateRE.FindStringSubmatch(s); m != nil {
		a, _ := strconv.Atoi(m[1])
		b, _ := strconv.Atoi(m[2])
		y := normalizeYear(m[3])

		switch p.PreferredFormat {
		case DateFormatDMY:
			return time.Date(y, time.Month(b), a, 0, 0, 0, 0, time.UTC), nil
		default: // MDY
			return time.Date(y, time.Month(a), b, 0, 0, 0, 0, time.UTC), nil
		}
	}

	// Dot separated (European)
	if m := dotDateRE.FindStringSubmatch(s); m != nil {
		d, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		y := normalizeYear(m[3])
		return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse date: %q", s)
}

// IsAmbiguous returns true if the date string could be interpreted as both MDY and DMY.
func IsAmbiguous(s string) bool {
	m := slashDateRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	return a <= 12 && b <= 12 && a != b
}

func normalizeYear(s string) int {
	y, _ := strconv.Atoi(s)
	if y < 100 {
		if y > 70 {
			y += 1900
		} else {
			y += 2000
		}
	}
	return y
}
