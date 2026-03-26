package context

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// GetInt64 is function to get some int64 value from gin context
func GetInt64(c *gin.Context, key string) (int64, bool) {
	if iv, ok := c.Get(key); !ok {
		return 0, false
	} else if v, ok := iv.(int64); !ok {
		return 0, false
	} else {
		return v, true
	}
}

// GetUint64 is function to get some uint64 value from gin context
func GetUint64(c *gin.Context, key string) (uint64, bool) {
	if iv, ok := c.Get(key); !ok {
		return 0, false
	} else if v, ok := iv.(uint64); !ok {
		return 0, false
	} else {
		return v, true
	}
}

// GetString is function to get some string value from gin context
func GetString(c *gin.Context, key string) (string, bool) {
	if iv, ok := c.Get(key); !ok {
		return "", false
	} else if v, ok := iv.(string); !ok {
		return "", false
	} else {
		return v, true
	}
}

// GetStringArray is function to get some string array value from gin context
func GetStringArray(c *gin.Context, key string) ([]string, bool) {
	if iv, ok := c.Get(key); !ok {
		return []string{}, false
	} else if v, ok := iv.([]string); !ok {
		return []string{}, false
	} else {
		return v, true
	}
}

func GetStringFromSession(c *gin.Context, key string) (string, bool) {
	session := sessions.Default(c)
	if iv := session.Get(key); iv == nil {
		return "", false
	} else if v, ok := iv.(string); !ok {
		return "", false
	} else {
		return v, true
	}
}
