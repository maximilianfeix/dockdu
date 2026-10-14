package report

import (
	"fmt"
	"unicode/utf8"
)

// Bytes formats a byte count in binary units with one decimal, or "unknown"
// for negative values, which the model uses for "not measured".
func Bytes(n int64) string {
	if n < 0 {
		return "unknown"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	v := float64(n)
	i := -1
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

// textWidth is the column limit for prose in the text report, so it reads
// on an 80-column terminal. Table rows and suggested commands are exempt:
// tables keep their columns aligned, and a command stays on one line so it
// can be copied.
const textWidth = 80

func runeLen(s string) int { return utf8.RuneCountInString(s) }

// wrapTokens joins tokens with sep and breaks lines only between tokens:
// the first line may hold firstWidth runes, later lines restWidth. A token
// longer than a line (a path) stays whole.
func wrapTokens(tokens []string, sep string, firstWidth, restWidth int) []string {
	var lines []string
	cur, width := "", firstWidth
	for _, t := range tokens {
		switch {
		case cur == "":
			cur = t
		case runeLen(cur)+runeLen(sep)+runeLen(t) <= width:
			cur += sep + t
		default:
			lines = append(lines, cur)
			cur, width = t, restWidth
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
