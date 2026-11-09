package ui

import (
	"strconv"
	"strings"
	"time"
)

func Humanize(v string) string {
	v = strings.ReplaceAll(v, "_", " ")
	if v == "" {
		return "unknown"
	}
	return strings.ToUpper(v[:1]) + v[1:]
}

func StateTone(state string) string {
	switch state {
	case "alive":
		return "success"
	case "flaky":
		return "warning"
	case "dead":
		return "danger"
	default:
		return "neutral"
	}
}

func OutcomeTone(outcome string) string {
	switch outcome {
	case "ok", "redirect_ok":
		return "success"
	case "blocked", "timeout", "server_error":
		return "warning"
	case "not_found", "gone", "dns_failure", "conn_refused", "tls_error", "redirect_loop", "other_error":
		return "danger"
	default:
		return "neutral"
	}
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Local().Format("2006-01-02 15:04")
}

func FormatTimeValue(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
}

func DerefString(v *string) string {
	if v == nil {
		return "—"
	}
	return *v
}

func DerefInt(v *int32) string {
	if v == nil {
		return "—"
	}
	return strconv.Itoa(int(*v))
}
