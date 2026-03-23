package app

import (
	"strings"
	"time"
)

func fmtDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.In(time.Local).Format("2006-01-02")
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
