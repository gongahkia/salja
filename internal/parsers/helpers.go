package parsers

import (
	"io"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
)

// findMissingColumns checks which required columns are absent from colMap.
func findMissingColumns(colMap map[string]int, required []string) []string {
	var missing []string
	for _, col := range required {
		if _, ok := colMap[col]; !ok {
			missing = append(missing, col)
		}
	}
	return missing
}

// transcodeReader wraps an io.Reader with automatic charset detection and UTF-8 transcoding.
func transcodeReader(r io.Reader) (io.Reader, error) {
	tr, _, err := salerr.TranscodeToUTF8(r)
	return tr, err
}

// parseAmbiguousDate uses locale-aware date parsing for ambiguous date strings.
var defaultDateParser = salerr.NewAmbiguousDateParser("")

func parseAmbiguousDate(s string) (time.Time, error) {
	return defaultDateParser.Parse(s)
}

// SetLocale updates the date parser with the given locale.
func SetLocale(locale string) {
	defaultDateParser = salerr.NewAmbiguousDateParser(locale)
}
