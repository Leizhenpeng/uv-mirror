package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
	JSDELIVR  bool
	CNPMJS    bool
	SizeLimit int64 // Maximum file size in bytes (2GB default)
	PREFIX    string
	AssetURL  string
	Whitelist []string // Whitelisted repositories (user/repo format)
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
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitHub Proxy</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #0d1117;
            color: #c9d1d9;
        }
        .container {
            text-align: center;
            padding: 40px 20px;
        }
        h1 {
            color: #58a6ff;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #8b949e;
            margin-bottom: 40px;
        }
        .input-group {
            margin: 20px 0;
        }
        input[type="text"] {
            width: 100%;
            max-width: 600px;
            padding: 12px 16px;
            border: 1px solid #30363d;
            border-radius: 6px;
            background-color: #21262d;
            color: #c9d1d9;
            font-size: 16px;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #58a6ff;
            box-shadow: 0 0 0 3px rgba(88, 166, 255, 0.1);
        }
        .btn {
            background-color: #238636;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            font-size: 16px;
            cursor: pointer;
            margin: 10px 5px;
            text-decoration: none;
            display: inline-block;
        }
        .btn:hover {
            background-color: #2ea043;
        }
        .example {
            margin-top: 40px;
            text-align: left;
            background-color: #161b22;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 20px;
        }
        .example h3 {
            color: #58a6ff;
            margin-top: 0;
        }
        .example ul {
            padding-left: 20px;
        }
        .example li {
            margin: 8px 0;
            word-break: break-all;
        }
        .example code {
            background-color: #21262d;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GitHub Proxy</h1>
        <p class="subtitle">GitHub 文件加速下载服务</p>
        
        <div class="input-group">
            <input type="text" id="url" placeholder="输入 GitHub 文件链接..." />
        </div>
        
        <button class="btn" onclick="goProxy()">加速下载</button>
        <button class="btn" onclick="copyProxy()">复制链接</button>
        
        <div class="example">
            <h3>支持的链接类型:</h3>
            <ul>
                <li>分支源码: <code>https://github.com/user/repo/archive/master.zip</code></li>
                <li>Release 源码: <code>https://github.com/user/repo/archive/v1.0.0.tar.gz</code></li>
                <li>Release 文件: <code>https://github.com/user/repo/releases/download/v1.0.0/file.zip</code></li>
                <li>分支文件: <code>https://github.com/user/repo/blob/master/README.md</code></li>
                <li>Raw 文件: <code>https://raw.githubusercontent.com/user/repo/master/file.txt</code></li>
                <li>Gist 文件: <code>https://gist.githubusercontent.com/user/id/raw/file.py</code></li>
            </ul>
            
            <h3>Git Clone 加速:</h3>
            <p>将原链接中的 <code>github.com</code> 替换为当前域名即可:</p>
            <code>git clone https://当前域名/user/repo.git</code>
        </div>
    </div>

    <script>
        function goProxy() {
            const url = document.getElementById('url').value.trim();
            if (!url) {
                alert('请输入 GitHub 链接');
                return;
            }
            
            let proxyUrl;
            if (url.startsWith('https://') || url.startsWith('http://')) {
                proxyUrl = location.origin + '/' + url;
            } else {
                proxyUrl = location.origin + '/https://' + url;
            }
            
            window.open(proxyUrl, '_blank');
        }
        
        function copyProxy() {
            const url = document.getElementById('url').value.trim();
            if (!url) {
                alert('请输入 GitHub 链接');
                return;
            }
            
            let proxyUrl;
            if (url.startsWith('https://') || url.startsWith('http://')) {
                proxyUrl = location.origin + '/' + url;
            } else {
                proxyUrl = location.origin + '/https://' + url;
            }
            
            navigator.clipboard.writeText(proxyUrl).then(() => {
                alert('代理链接已复制到剪贴板');
            }).catch(() => {
                prompt('复制链接:', proxyUrl);
            });
        }
        
        // Handle Enter key
        document.getElementById('url').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                goProxy();
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

	// For jsdelivr acceleration (optional)
	if p.config.JSDELIVR && p.canUseJsdelivr(targetURL) {
		jsdelivrURL := p.convertToJsdelivr(targetURL)
		if jsdelivrURL != "" {
			http.Redirect(w, r, jsdelivrURL, http.StatusTemporaryRedirect)
			return
		}
	}

	// For cnpmjs acceleration (optional)
	if p.config.CNPMJS && p.canUseCnpmjs(targetURL) {
		cnpmjsURL := p.convertToCnpmjs(targetURL)
		if cnpmjsURL != "" {
			http.Redirect(w, r, cnpmjsURL, http.StatusTemporaryRedirect)
			return
		}
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
	// If no whitelist is configured, allow all repositories
	if len(p.config.Whitelist) == 0 {
		return true
	}

	// Extract repository path from URL
	repoPath := p.extractRepositoryPath(targetURL)
	if repoPath == "" {
		return false
	}

	// Check if repository is in whitelist
	for _, whitelistedRepo := range p.config.Whitelist {
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

// Check if URL can use jsdelivr
func (p *ProxyHandler) canUseJsdelivr(targetURL *url.URL) bool {
	// jsdelivr supports github.com and raw.githubusercontent.com
	return targetURL.Host == GitHubDomain || targetURL.Host == GitHubRawDomain
}

// Convert to jsdelivr URL
func (p *ProxyHandler) convertToJsdelivr(targetURL *url.URL) string {
	path := targetURL.Path

	if targetURL.Host == GitHubRawDomain {
		// raw.githubusercontent.com/user/repo/branch/file -> cdn.jsdelivr.net/gh/user/repo@branch/file
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) >= 3 {
			user := parts[0]
			repo := parts[1]
			branch := parts[2]
			file := strings.Join(parts[3:], "/")
			return fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s", user, repo, branch, file)
		}
	} else if targetURL.Host == GitHubDomain {
		// github.com/user/repo/blob/branch/file -> cdn.jsdelivr.net/gh/user/repo@branch/file
		if strings.Contains(path, "/blob/") {
			re := regexp.MustCompile(`^/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$`)
			matches := re.FindStringSubmatch(path)
			if len(matches) == 5 {
				user := matches[1]
				repo := matches[2]
				branch := matches[3]
				file := matches[4]
				return fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s", user, repo, branch, file)
			}
		}
	}

	return ""
}

// Check if URL can use cnpmjs
func (p *ProxyHandler) canUseCnpmjs(targetURL *url.URL) bool {
	// This is just an example, implement based on cnpmjs capabilities
	return false
}

// Convert to cnpmjs URL
func (p *ProxyHandler) convertToCnpmjs(targetURL *url.URL) string {
	// Implement cnpmjs conversion logic if needed
	return ""
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
		JSDELIVR:  getEnv("JSDELIVR", "false") == "true",
		CNPMJS:    getEnv("CNPMJS", "false") == "true",
		SizeLimit: 2 * 1024 * 1024 * 1024, // 2GB default
		PREFIX:    getEnv("PREFIX", "/"),
		AssetURL:  getEnv("ASSET_URL", ""),
		Whitelist: parseWhitelist(getEnv("WHITELIST", "")),
	}

	// Create proxy handler
	handler := NewProxyHandler(config)

	// Setup HTTP server
	http.Handle("/", handler)

	log.Printf("GitHub Proxy server starting on port %s", config.Port)
	log.Printf("JSDELIVR acceleration: %v", config.JSDELIVR)
	log.Printf("CNPMJS acceleration: %v", config.CNPMJS)
	if len(config.Whitelist) > 0 {
		log.Printf("Whitelist enabled with %d repositories: %v", len(config.Whitelist), config.Whitelist)
	} else {
		log.Printf("Whitelist disabled - all repositories allowed")
	}

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

// Parse whitelist from comma-separated string
func parseWhitelist(whitelistStr string) []string {
	if whitelistStr == "" {
		return []string{}
	}

	var whitelist []string
	parts := strings.Split(whitelistStr, ",")
	for _, part := range parts {
		repo := strings.TrimSpace(part)
		if repo != "" {
			whitelist = append(whitelist, repo)
		}
	}

	return whitelist
}
