package report

import (
	"fmt"
	"strings"
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

// units are the unit words Bytes prints after a number.
var units = map[string]bool{"B": true, "KiB": true, "MiB": true, "GiB": true, "TiB": true, "PiB": true}

// words splits prose into wrap tokens, keeping a number and its unit
// ("140.7 MiB", "0 B") in one token so a line never ends with a bare number.
func words(text string) []string {
	f := strings.Fields(text)
	out := make([]string, 0, len(f))
	for _, w := range f {
		unit := strings.TrimRight(w, ".,;:)")
		if n := len(out); n > 0 && units[unit] && endsWithDigit(out[n-1]) {
			out[n-1] += " " + w
			continue
		}
		out = append(out, w)
	}
	return out
}

func endsWithDigit(s string) bool { return s != "" && s[len(s)-1] >= '0' && s[len(s)-1] <= '9' }
