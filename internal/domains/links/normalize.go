package links

import (
	"crypto/sha256"
	"link_or_die/internal/apperror"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var trackingParams = map[string]struct{}{
	"utm_source":   {},
	"utm_medium":   {},
	"utm_campaign": {},
	"utm_term":     {},
	"utm_content":  {},
	"fbclid":       {},
	"gclid":        {},
	"dclid":        {},
	"msclkid":      {},
	"mc_cid":       {},
	"mc_eid":       {},
	"_ga":          {},
}

func Normalize(raw string) (string, string, []byte, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return "", "", nil, apperror.ErrInvalidURL
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", "", nil, apperror.ErrInvalidURL
	}

	u.Scheme = strings.ToLower(u.Scheme)

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", nil, apperror.ErrInvalidURL
	}

	if u.Opaque != "" || u.Host == "" {
		return "", "", nil, apperror.ErrInvalidURL
	}

	if u.User != nil {
		return "", "", nil, apperror.ErrInvalidURL
	}

	hostname := strings.ToLower(u.Hostname())
	if hostname == "" {
		return "", "", nil, apperror.ErrInvalidURL
	}

	if !strings.Contains(hostname, ":") {
		hostname = strings.TrimSuffix(hostname, ".")
	}

	if hostname == "" {
		return "", "", nil, apperror.ErrInvalidURL
	}

	port := u.Port()

	if port != "" {
		n, parseErr := strconv.Atoi(port)
		if parseErr != nil || n < 0 || n > 65535 {
			return "", "", nil, apperror.ErrInvalidURL
		}

		// Remove default ports.
		if (u.Scheme == "http" && n == 80) ||
			(u.Scheme == "https" && n == 443) {
			port = ""
		}
	}

	switch {
	case port != "":
		u.Host = net.JoinHostPort(hostname, port)

	case strings.Contains(hostname, ":"):
		u.Host = "[" + hostname + "]"

	default:
		u.Host = hostname
	}

	host := hostname

	u.Fragment = ""
	u.RawFragment = ""
	// normalise empty path to /
	if u.Path == "" {
		u.Path = "/"
		u.RawPath = ""
	}

	// Remove tracking parameters.
	q, queryErr := url.ParseQuery(u.RawQuery)
	if queryErr != nil {
		return "", "", nil, apperror.ErrInvalidURL
	}

	for key := range q {
		if _, ok := trackingParams[strings.ToLower(key)]; ok {
			q.Del(key)
		}
	}

	u.RawQuery = q.Encode()
	u.ForceQuery = false

	normalized := u.String()

	sum := sha256.Sum256([]byte(normalized))
	hash := sum[:]

	return normalized, host, hash, nil
}
