package parsers

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
