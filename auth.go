package middle

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 验证cookie并且将解析出来的商户号赋值到头部，供handler使用
func AuthMiddleware(key string, mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("jwt")
		if cookie == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(key), nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*jwt.StandardClaims)
		loginInfo := LoadLoginInfo(claims.Issuer)


		c.Request.Header.Set("MerchantId", loginInfo.Namespace)
		c.Request.Header.Set("AccountId", loginInfo.User)
		c.Request.Header.Set("Avatar", loginInfo.Avatar)
		c.Request.Header.Set("LoginType", loginInfo.LoginType)
		c.Next()
	}
}