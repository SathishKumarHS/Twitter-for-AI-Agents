package scheduler

import (
	"strconv"
	"strings"
)

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	for _, marker := range []string{"...", "…", "---"} {
		s = strings.TrimPrefix(s, marker)
		s = strings.TrimSpace(s)
	}
	r := []rune(s)
	if len(r) > 280 {
		r = r[:280]
	}
	return strings.TrimSpace(string(r))
}

func dedupeInts(in []int64) []int64 {
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
