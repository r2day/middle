package middle

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	// 普通用户
	LoginTypeNormal = "normal"
	// 管理员
	LoginTypeAdmin = "admin"
	// 加盟
	LoginTypeJoin = "join"
	// 退出地址
	SignOutPath = "/v1/auth/merchant/signout"
)

// AuthMiddleware 验证cookie并且将解析出来的商户号赋值到头部，供handler使用
func AuthMiddleware(key []byte, mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("jwt")
		if cookie == "" {
			log.Error("cookie name as jwt no found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			log.WithField("message", "parse claims failed").Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*jwt.StandardClaims)
		loginInfo, err := LoadLoginInfo(claims.Issuer)
		if err != nil {
			log.WithField("message", "LoadLoginInfo failed").Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 写入解析客户的jwt token后得到的数据
		c.Request.Header.Set("MerchantId", loginInfo.Namespace)
		c.Request.Header.Set("AccountId", loginInfo.AccountId)
		c.Request.Header.Set("UserId", loginInfo.UserId)
		c.Request.Header.Set("UserName", loginInfo.UserName)
		c.Request.Header.Set("Avatar", loginInfo.Avatar)
		c.Request.Header.Set("LoginType", loginInfo.LoginType)
		c.Request.Header.Set("LoginLevel", loginInfo.LoginLevel)
		// 日志检查各项数据是否回传正常
		c.Next()
	}
}

// AuthWithWhitelistMiddleware 验证cookie并且将解析出来的商户号赋值到头部，供handler使用
func AuthWithWhitelistMiddleware(key []byte, skipPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("jwt")
		if cookie == "" {
			if c.FullPath() == skipPath {
				// 日志检查各项数据是否回传正常
				log.WithField("path", c.FullPath()).
					Warning("cookie is empty, but it's not require for this path, so continue")
				c.Next()
				return
			}
			log.Error("cookie name as jwt no found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			log.WithField("message", "parse claims failed").Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*jwt.StandardClaims)
		loginInfo, err := LoadLoginInfo(claims.Issuer)
		if err != nil {
			log.WithField("message", "LoadLoginInfo failed").Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 写入解析客户的jwt token后得到的数据
		c.Request.Header.Set("MerchantId", loginInfo.Namespace)
		c.Request.Header.Set("AccountId", loginInfo.AccountId)
		c.Request.Header.Set("UserId", loginInfo.UserId)
		c.Request.Header.Set("UserName", loginInfo.UserName)
		c.Request.Header.Set("Avatar", loginInfo.Avatar)
		c.Request.Header.Set("LoginType", loginInfo.LoginType)
		c.Request.Header.Set("LoginLevel", loginInfo.LoginLevel)
		// 日志检查各项数据是否回传正常
		c.Next()
	}
}
