package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	// Default port
	DefaultPort = "8080"

	// GitHub domains that need to be proxied
	GitHubDomain            = "github.com"
	GitHubAPIDomain         = "api.github.com"
	GitHubRawDomain         = "raw.githubusercontent.com"
	GitHubGistDomain        = "gist.githubusercontent.com"
	GitHubAssetsDomain      = "github.githubassets.com"
	GitHubAvatarsDomain     = "avatars.githubusercontent.com"
	GitHubCamoDomain        = "camo.githubusercontent.com"
	GitHubUserContentDomain = "user-images.githubusercontent.com"
)

// Config holds the configuration for the proxy
type Config struct {
	Port      string
	SizeLimit int64 // Maximum file size in bytes (2GB default)
}

// ProxyHandler handles all proxy requests
type ProxyHandler struct {
	config *Config
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(config *Config) *ProxyHandler {
	return &ProxyHandler{config: config}
}

// Main handler for all requests
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle robots.txt
	if r.URL.Path == "/robots.txt" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
		return
	}

	// Handle homepage
	if r.URL.Path == "/" {
		p.handleHomepage(w, r)
		return
	}

	// Handle proxy requests
	p.handleProxy(w, r)
}

// Handle homepage with input form
func (p *ProxyHandler) handleHomepage(w http.ResponseWriter, r *http.Request) {
	html := `GitHub Proxy - 一个简单的GitHub文件加速服务`

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(html))
}

// Handle proxy requests
func (p *ProxyHandler) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Extract target URL from path
	path := strings.TrimPrefix(r.URL.Path, "/")

	// Handle the case where URL starts with https:/ instead of https://
	if strings.HasPrefix(path, "https:/") && !strings.HasPrefix(path, "https://") {
		path = "https://" + strings.TrimPrefix(path, "https:/")
	} else if strings.HasPrefix(path, "http:/") && !strings.HasPrefix(path, "http://") {
		path = "http://" + strings.TrimPrefix(path, "http:/")
	}

	// If path doesn't start with http, add https://
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = "https://" + path
	}

	// Parse target URL
	targetURL, err := url.Parse(path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Check if it's a GitHub-related domain
	if !p.isGitHubDomain(targetURL.Host) {
		http.Error(w, "Only GitHub domains are supported", http.StatusForbidden)
		return
	}

	// Check if repository is whitelisted
	if !p.isRepositoryWhitelisted(targetURL) {
		http.Error(w, "Repository not in whitelist", http.StatusForbidden)
		return
	}

	// Handle git clone requests
	if p.isGitCloneRequest(r) {
		p.handleGitClone(w, r, targetURL)
		return
	}

	// Proxy the request
	p.proxyRequest(w, r, targetURL)
}

// Check if domain is GitHub-related
func (p *ProxyHandler) isGitHubDomain(host string) bool {
	githubDomains := []string{
		GitHubDomain,
		GitHubAPIDomain,
		GitHubRawDomain,
		GitHubGistDomain,
		GitHubAssetsDomain,
		GitHubAvatarsDomain,
		GitHubCamoDomain,
		GitHubUserContentDomain,
	}

	for _, domain := range githubDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

// Check if repository is in whitelist
func (p *ProxyHandler) isRepositoryWhitelisted(targetURL *url.URL) bool {
	// Hardcoded whitelist - add your allowed repositories here
	whitelist := []string{
		"astral-sh/uv",
		// Add more repositories as needed
	}

	// If whitelist is empty, allow all repositories
	if len(whitelist) == 0 {
		return true
	}

	// Extract repository path from URL
	repoPath := p.extractRepositoryPath(targetURL)
	if repoPath == "" {
		return false
	}

	// Check if repository is in whitelist
	for _, whitelistedRepo := range whitelist {
		if strings.EqualFold(repoPath, whitelistedRepo) {
			return true
		}
	}

	return false
}

// Extract repository path (user/repo) from GitHub URL
func (p *ProxyHandler) extractRepositoryPath(targetURL *url.URL) string {
	path := strings.Trim(targetURL.Path, "/")

	// Handle different GitHub domain patterns
	switch targetURL.Host {
	case GitHubDomain:
		// github.com/user/repo/...
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	case GitHubRawDomain:
		// raw.githubusercontent.com/user/repo/...
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	case GitHubAPIDomain:
		// api.github.com/repos/user/repo/...
		if strings.HasPrefix(path, "repos/") {
			parts := strings.Split(strings.TrimPrefix(path, "repos/"), "/")
			if len(parts) >= 2 {
				return parts[0] + "/" + parts[1]
			}
		}
	case GitHubGistDomain:
		// gist.githubusercontent.com/user/gist_id/...
		parts := strings.Split(path, "/")
		if len(parts) >= 1 {
			return "gist:" + parts[0] // Special format for gists
		}
	}

	return ""
}

// Check if request is for git clone
func (p *ProxyHandler) isGitCloneRequest(r *http.Request) bool {
	userAgent := r.Header.Get("User-Agent")
	return strings.Contains(userAgent, "git/")
}

// Handle git clone requests
func (p *ProxyHandler) handleGitClone(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
	// For git clone, we need to proxy the request directly
	p.proxyRequest(w, r, targetURL)
}

// Proxy the request to target URL
func (p *ProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) {
	// Create new request
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		// Skip certain headers
		if name == "Host" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Set host header
	proxyReq.Host = targetURL.Host
	proxyReq.Header.Set("Host", targetURL.Host)

	// Set user agent if not present
	if proxyReq.Header.Get("User-Agent") == "" {
		proxyReq.Header.Set("User-Agent", "gh-proxy-go/1.0")
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch from target", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Check file size limit
	if p.config.SizeLimit > 0 {
		if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
			if size := parseContentLength(contentLength); size > p.config.SizeLimit {
				http.Redirect(w, r, targetURL.String(), http.StatusTemporaryRedirect)
				return
			}
		}
	}

	// Copy response headers
	for name, values := range resp.Header {
		// Skip certain headers
		if name == "Content-Encoding" || name == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

// Parse content length
func parseContentLength(s string) int64 {
	if s == "" {
		return 0
	}
	var size int64
	fmt.Sscanf(s, "%d", &size)
	return size
}

func main() {
	// Get configuration from environment variables
	config := &Config{
		Port:      getEnv("PORT", DefaultPort),
		SizeLimit: 2 * 1024 * 1024 * 1024, // 2GB default
	}

	// Create proxy handler
	handler := NewProxyHandler(config)

	// Setup HTTP server
	http.Handle("/", handler)

	log.Printf("GitHub Proxy server starting on port %s", config.Port)

	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
