package server

import "topography/v2/internal/dataset"

const (
	HTML_FILES = "min/html/*.html"
	SCCMP_FILE = "min/security/seccomp.txt"

	HTML_CACHE    = "public, max-age=3600, immutable"     // 1 hour
	DEFAULT_CACHE = "public, max-age=3600, immutable"     // 1 hour
	STATIC_CACHE  = "public, max-age=86400, immutable"    // 1 day
	TOPO_CACHE    = "public, max-age=31536000, immutable" // 1 year

	MAX_RESOLUTION = 4096
	MIN_RESOLUTION = 512
	STEP_VALUE     = 512
	TARGET_ORIGIN  = dataset.SW_ORIGIN
)
