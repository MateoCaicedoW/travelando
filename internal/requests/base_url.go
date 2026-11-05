package requests

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func appBaseURL(r *http.Request) (string, error) {
	base := strings.TrimSpace(os.Getenv("BASE_URL"))
	if base == "" {
		host, _, err := net.SplitHostPort(r.Host)
		if err == nil && (host == "localhost" || net.ParseIP(host).IsLoopback()) {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			base = scheme + "://" + r.Host
		} else {
			return "", fmt.Errorf("BASE_URL is required outside localhost")
		}
	}
	parsed, err := url.Parse(base)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("BASE_URL must be an absolute HTTP(S) URL")
	}
	if os.Getenv("GO_ENV") == "production" && parsed.Scheme != "https" {
		return "", fmt.Errorf("BASE_URL must use HTTPS in production")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}
