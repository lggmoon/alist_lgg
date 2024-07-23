package middlewares

import (
	"crypto/subtle"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Auth is a middleware that checks if the user is logged in.
// if token is empty, set user to guest
func Auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if subtle.ConstantTimeCompare([]byte(token), []byte(setting.GetStr(conf.Token))) == 1 {
		admin, err := op.GetAdmin()
		if err != nil {
			common.ErrorResp(c, err, 500)
			c.Abort()
			return
		}
		c.Set("user", admin)
		log.Debugf("use admin token: %+v", admin)
		c.Next()
		return
	}
	if token == "" {
		guest, err := op.GetGuest()
		if err != nil {
			common.ErrorResp(c, err, 500)
			c.Abort()
			return
		}
		if guest.Disabled {
			common.ErrorStrResp(c, "Guest user is disabled, login please", 401)
			c.Abort()
			return
		}
		c.Set("user", guest)
		log.Debugf("use empty token: %+v", guest)
		c.Next()
		return
	}
	userClaims, err := common.ParseToken(token)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	user, err := op.GetUserByName(userClaims.Username)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	// validate password timestamp
	if userClaims.PwdTS != user.PwdTS {
		common.ErrorStrResp(c, "Password has been changed, login please", 401)
		c.Abort()
		return
	}
	if user.Disabled {
		common.ErrorStrResp(c, "Current user is disabled, replace please", 401)
		c.Abort()
		return
	}
	c.Set("user", user)
	log.Debugf("use login token: %+v", user)
	c.Next()
}

func Authn(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if subtle.ConstantTimeCompare([]byte(token), []byte(setting.GetStr(conf.Token))) == 1 {
		admin, err := op.GetAdmin()
		if err != nil {
			common.ErrorResp(c, err, 500)
			c.Abort()
			return
		}
		c.Set("user", admin)
		log.Debugf("use admin token: %+v", admin)
		c.Next()
		return
	}
	if token == "" {
		guest, err := op.GetGuest()
		if err != nil {
			common.ErrorResp(c, err, 500)
			c.Abort()
			return
		}
		c.Set("user", guest)
		log.Debugf("use empty token: %+v", guest)
		c.Next()
		return
	}
	userClaims, err := common.ParseToken(token)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	user, err := op.GetUserByName(userClaims.Username)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	// validate password timestamp
	if userClaims.PwdTS != user.PwdTS {
		common.ErrorStrResp(c, "Password has been changed, login please", 401)
		c.Abort()
		return
	}
	if user.Disabled {
		common.ErrorStrResp(c, "Current user is disabled, replace please", 401)
		c.Abort()
		return
	}
	c.Set("user", user)
	log.Debugf("use login token: %+v", user)
	c.Next()
}

func _genernal_pass(url string) bool {
	if strings.HasPrefix(url, "/@manage/storages") {
		return true
	} else if strings.HasPrefix(url, "/api/admin/storage") {
		return true
	} else if strings.HasPrefix(url, "/api/admin/driver") {
		return true
	} else if strings.HasPrefix(url, "/api/admin/task") {
		return true
	}

	// str_array := []string{}
	// for _, element := range str_array {
	// 	if url == element {
	// 		return true
	// 	}
	// }
	return false
}

func AuthAdmin(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	pass := user.IsGeneral() && _genernal_pass(c.Request.URL.Path)
	if !pass && !user.IsAdmin() {
		common.ErrorStrResp(c, "You are not an admin", 403)
		c.Abort()
	} else {
		c.Next()
	}
}
