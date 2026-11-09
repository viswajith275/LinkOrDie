package checker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"link_or_die/internal/config"
	"link_or_die/internal/domains/links"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

var (
	errTooManyRedirects = errors.New("too many redirects")
	errUnsafeRedirect   = errors.New("unsafe redirect")
)

// Result of check
type Result struct {
	Outcome    links.Outcome
	StatusCode *int32
	FinalURL   string
	Method     string
	Elapsed    time.Duration
	Error      string
	Canceled   bool
}

// Checker config
type Checker struct {
	client    *http.Client
	timeout   time.Duration
	maxBody   int64
	userAgent string
}

func NewChecker(cfg *config.Config) *Checker {
	dialer := &net.Dialer{
		Timeout:   cfg.Timeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                  nil, // never inherit HTTP_PROXY
		DialContext:            safeDialContext(dialer),
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:    cfg.Timeout,
		ResponseHeaderTimeout:  cfg.Timeout,
		ExpectContinueTimeout:  time.Second,
		MaxResponseHeaderBytes: 64 << 10,
		IdleConnTimeout:        90 * time.Second,
		DisableCompression:     true,
		ForceAttemptHTTP2:      true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return errTooManyRedirects
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errUnsafeRedirect
			}
			if req.URL.User != nil {
				return errUnsafeRedirect
			}
			return nil
		},
	}

	return &Checker{
		client:    client,
		timeout:   cfg.Timeout,
		maxBody:   cfg.MaxBodyBytes,
		userAgent: cfg.UserAgent,
	}
}

func (c *Checker) Check(ctx context.Context, URL string) Result {
	start := time.Now()

	result := Result{
		FinalURL: URL,
		Method:   http.MethodGet,
	}

	u, err := url.Parse(strings.TrimSpace(URL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		result.Outcome = links.OutcomeOtherError
		result.Error = "invalid url"
		result.Elapsed = time.Since(start)
		return result
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		result.Outcome = links.OutcomeOtherError
		result.Error = err.Error()
		result.Elapsed = time.Since(start)
		return result
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.client.Do(req)
	result.Elapsed = time.Since(start)

	if err != nil {
		result.Outcome = classifyError(err)
		result.Error = err.Error()
		result.Canceled = errors.Is(err, context.Canceled) && ctx.Err() == context.Canceled
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = int32ptr(int32(resp.StatusCode))
	result.FinalURL = resp.Request.URL.String()
	result.Method = req.Method
	result.Outcome = classifyStatus(resp.StatusCode)

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, c.maxBody))

	return result
}

func int32ptr(v int32) *int32 {
	return &v
}

func isTLSError(err error) bool {
	var (
		certVerify *tls.CertificateVerificationError
		record     tls.RecordHeaderError
		unknown    x509.UnknownAuthorityError
		hostname   x509.HostnameError
		invalid    x509.CertificateInvalidError
	)

	return errors.As(err, &certVerify) ||
		errors.As(err, &record) ||
		errors.As(err, &unknown) ||
		errors.As(err, &hostname) ||
		errors.As(err, &invalid)
}

func classifyStatus(code int) links.Outcome {
	switch {
	case code >= 200 && code < 300:

		return links.OutcomeOK

	case code >= 300 && code < 400:

		return links.OutcomeRedirectOK

	case code == http.StatusUnauthorized,
		code == http.StatusForbidden,
		code == http.StatusTooManyRequests:

		return links.OutcomeBlocked

	case code == http.StatusNotFound:

		return links.OutcomeNotFound

	case code == http.StatusGone:

		return links.OutcomeGone

	case code >= 500 && code < 600:

		return links.OutcomeServerError

	default:

		return links.OutcomeOtherError
	}
}

func classifyError(err error) links.Outcome {
	switch {
	case errors.Is(err, errTooManyRedirects):
		return links.OutcomeRedirectLoop
	case errors.Is(err, errUnsafeRedirect):
		return links.OutcomeOtherError
	case errors.Is(err, context.DeadlineExceeded):
		return links.OutcomeTimeout
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return links.OutcomeDNSFailure
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return links.OutcomeConnRefused
	}

	if isTLSError(err) {
		return links.OutcomeTLSError
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return links.OutcomeTimeout
	}

	return links.OutcomeOtherError
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsInterfaceLocalMulticast() {
		return true
	}

	v4 := ip.To4()
	if v4 == nil {
		return false
	}

	switch {
	case v4[0] == 0:
		return true
	case v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127:
		return true
	case v4[0] == 192 && v4[1] == 0 && v4[2] == 0:
		return true
	case v4[0] == 198 && (v4[1] == 18 || v4[1] == 19):
		return true
	case v4[0] >= 240:
		return true
	}

	return false
}

func safeDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		if ip := net.ParseIP(host); ip != nil {
			if isBlockedIP(ip) {
				return nil, fmt.Errorf("blocked address: %s", ip)
			}
			return dialer.DialContext(ctx, network, addr)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}

		var lastErr error
		for _, ip := range ips {
			if isBlockedIP(ip.IP) {
				lastErr = fmt.Errorf("blocked address: %s", ip.IP)
				continue
			}

			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}

		if lastErr == nil {
			lastErr = fmt.Errorf("no usable address for %s", host)
		}
		return nil, lastErr
	}
}
