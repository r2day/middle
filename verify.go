package middle

import (
	"github.com/gin-gonic/gin"
	"github.com/r2day/db"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// 微信登陆token
const wxLoginKey = "wx_tokens_"

// VerifyTokenMiddleware 微信登陆token校验
func VerifyTokenMiddleware(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		tokenKeyName := wxLoginKey + token
		openid, err := db.RDB.HGet(c.Request.Context(), tokenKeyName, "openid").Result()
		if err != nil {
			// 当redis中没有黑名单，则继续完成下面的步骤
			// 表示不需要在sidebar 中隐藏该项
			log.WithField("openId", openid).Error(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		log.WithField("token", token).WithField("openid", openid).Warning("=============")
		if openid == "" {
			// 当redis中没有黑名单，则继续完成下面的步骤
			// 表示不需要在sidebar 中隐藏该项
			log.WithField("openId", openid).Error(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no open id"})
			return
		}
		// 写入解析客户的jwt token后得到的数据
		c.Request.Header.Set("WX_OPEN_ID", openid)

		c.Next()
	}
}
