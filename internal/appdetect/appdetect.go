package appdetect

// DetectedApp represents a locally detected calendar/task app.
type DetectedApp struct {
	Name       string   // human-readable name
	FormatName string   // registry format key
	DataPaths  []string // known data file paths (if found)
	Installed  bool
}
