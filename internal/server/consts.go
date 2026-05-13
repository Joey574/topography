package server

const (
	HTML_FILES = "min/html/*.html"
	SCCMP_FILE = "min/security/seccomp.txt"

	HTML_CACHE    = "public, max-age=3600, immutable"     // 1 hour
	DEFAULT_CACHE = "public, max-age=3600, immutable"     // 1 hour
	STATIC_CACHE  = "public, max-age=2592000, immutable"  // 1 month
	TOPO_CACHE    = "public, max-age=31536000, immutable" // 1 year
)
