package helpers

const (
	HeaderContentType   = "Content-Type"
	MIMEApplicationJSON = "application/json"
	HeaderXCache        = "X-Cache"
	CacheHit            = "HIT"
	CacheMiss           = "MISS"

	DefaultFeedLimit = 50
	MaxFeedLimit     = 200
)

var CachedFeedLimits = []int{20, 50, 100}

// IsCachedFeedLimit reports whether a feed limit is served from / written to Redis.
func IsCachedFeedLimit(limit int) bool {
	for _, l := range CachedFeedLimits {
		if l == limit {
			return true
		}
	}
	return false
}
