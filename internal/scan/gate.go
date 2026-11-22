package scan

var severityRank = map[string]int{"info": 1, "warning": 2, "error": 3}

// AtOrAbove counts the findings whose severity is at or above min. An unknown
// or empty min counts nothing, so a gate that was not asked for never trips.
func (r RuleReport) AtOrAbove(min string) int {
	threshold, ok := severityRank[min]
	if !ok {
		return 0
	}
	n := 0
	for _, f := range r.Findings {
		if severityRank[f.Severity] >= threshold {
			n++
		}
	}
	return n
}
