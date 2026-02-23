package cache

type CacheState bool

var (
	CacheStateActive      CacheState = true
	CacheStateExpired     CacheState = false
	CacheStateNonExistant CacheState = false
)
