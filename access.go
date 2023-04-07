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
)

// AccessMiddleware 验证cookie并且将解析出来的账号
// 通过账号获取角色
// 通过角色判断其是否具有该api的访问权限
// 用户登陆完成后会将权限配置信息写入 redis 数据库完成
// 通过hget api/path/ role boolean
// 
func AccessMiddleware(key []byte, redisAddr string) gin.HandlerFunc {
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
		loginInfo := LoadLoginInfo(claims.Issuer)

		// 写入解析客户的jwt token后得到的数据
		c.Request.Header.Set("MerchantId", loginInfo.Namespace)
		c.Request.Header.Set("AccountId", loginInfo.AccountId)
		c.Request.Header.Set("UserId", loginInfo.UserId)
		c.Request.Header.Set("UserName", loginInfo.UserName)
		c.Request.Header.Set("Avatar", loginInfo.Avatar)
		c.Request.Header.Set("LoginType", loginInfo.LoginType)
		// 查询数据库
		roles := []string{"admin", "test"}
		// 检测角色是否有权限
		isAccess := CanAccess(redisAddr, c.FullPath(), roles)
		if !isAccess {
			log.WithField("message", "access denied").Error(err)
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		c.Next()
	}
}


func CanAccess(redisAddr string, path string, roles []string) bool {
	access := false
	rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	// 仅进行路径的请求访问权限校验
	for _, role := range roles {
		val, err := rdb.Hget(ctx, path, role).Result()
		if err != nil {
			log.WithField("message", "no acceptable ").Error(err)
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}
		// is true
		// 如果有一个角色是true 则代表其可以访问
		if bool(val) {
			return true
		}
	}
	return false
}