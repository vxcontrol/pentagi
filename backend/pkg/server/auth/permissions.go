package auth

import (
	"fmt"

	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
)

func getPrms(c *gin.Context) ([]string, error) {
	prms := c.GetStringSlice("prm")
	if len(prms) == 0 {
		return nil, fmt.Errorf("privileges are not set")
	}
	return prms, nil
}

func PrivilegesRequired(privs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.IsAborted() {
			return
		}

		prms, err := getPrms(c)
		if err != nil {
			response.Error(c, response.ErrPrivilegesRequired, err)
			c.Abort()
			return
		}

		for _, priv := range append([]string{}, privs...) {
			if !LookupPerm(prms, priv) {
				response.Error(c, response.ErrPrivilegesRequired, fmt.Errorf("'%s' is not set", priv))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func LookupPerm(prm []string, perm string) bool {
	for _, p := range prm {
		if p == perm {
			return true
		}
	}
	return false
}
