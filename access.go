package middle

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/r2day/db"
	log "github.com/sirupsen/logrus"
)

const (
	// AccessKeyPrefix redis 前缀
	AccessKeyPrefix = "access_key_prefix"
)

// AccessMiddleware 验证cookie并且将解析出来的账号
// 通过账号获取角色
// 通过角色判断其是否具有该api的访问权限
// 用户登陆完成后会将权限配置信息写入 redis 数据库完成
// 通过hget api/path/ role boolean
func AccessMiddleware(key []byte) gin.HandlerFunc {
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
		// 检测角色是否有权限
		isAccess := CanAccess(c.Request.Context(), c.FullPath(), loginInfo.AccountId)
		if !isAccess {
			log.WithField("message", "access denied").Error(err)
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		// 检测账号是否有操作权限
		isCanDo := CanDo(c.Request.Context(), c.FullPath(), loginInfo.AccountId, c.Request.Method)
		if !isCanDo {
			log.WithField("message", "operation denied").Error(err)
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		c.Next()
	}
}

// CanAccess 是否允许访问
func CanAccess(ctx context.Context, path string, accountID string) bool {

	// 该账号下面的用户角色
	keyPrefix := AccessKeyPrefix + "_" + accountID
	keyForRoles := keyPrefix + "_" + "roles"
	roles, err := db.RDB.SMembers(ctx, keyForRoles).Result()
	if err != nil {
		log.WithField("keyForRoles is no found", keyForRoles).Error(err)
		return false
	}

	// 仅进行路径的请求访问权限校验
	key := AccessKeyPrefix + "_" + accountID + "_" + "path_access"
	for _, role := range roles {
		//key := AccessKeyPrefix + "_" + accountID + "_" + path
		pathWithRole := path + "_" + role
		val, err := db.RDB.HGet(ctx, key, pathWithRole).Result()
		if err != nil {
			// 可以忽略该日志
			// 一般情况下仅角色匹配到path即可访问
			// 其他角色大部分会走该逻辑，因此将日志类别定义为debug
			log.WithField("message", "call db.RDB.HGet failed").
				WithField("val", val).
				WithField("path", path).
				WithField("key", key).
				WithField("role", role).
				Debug(err)
			continue
		}
		// is true
		// 如果有一个角色是true 则代表其可以访问
		boolValue, err := strconv.ParseBool(val)
		if err != nil {
			// 可以忽略该日志
			// 一般情况下仅角色匹配到path即可访问
			// 其他角色大部分会走该逻辑，因此将日志类别定义为debug
			log.WithField("message", "call strconv.ParseBool failed").
				WithField("path", path).WithField("key", key).
				WithField("boolValue", boolValue).Debug(err)
			continue
		}

		if boolValue {
			return true
		}
	}
	return false
}

// CanDo 是否允许操作
func CanDo(ctx context.Context, path string, accountID string, method string) bool {

	// 该账号下面的用户角色
	keyOperation := AccessKeyPrefix + "_" + accountID + "_operations"

	//key := AccessKeyPrefix + "_" + accountID + "_" + path
	pathWithMethod := path + "/" + method
	val, err := db.RDB.HGet(ctx, keyOperation, pathWithMethod).Result()
	if err != nil {
		// 可以忽略该日志
		// 一般情况下仅角色匹配到path即可访问
		// 其他角色大部分会走该逻辑，因此将日志类别定义为debug
		log.WithField("message", "call db.RDB.HGet failed").
			WithField("val", val).
			WithField("path", path).
			WithField("pathWithMethod", pathWithMethod).
			Debug(err)
		return false
	}
	// is true
	// 如果有一个角色是true 则代表其可以访问
	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		// 可以忽略该日志
		// 一般情况下仅角色匹配到path即可访问
		// 其他角色大部分会走该逻辑，因此将日志类别定义为debug
		log.WithField("message", "call strconv.ParseBool failed").
			WithField("path", path).WithField("key", key).
			WithField("boolValue", boolValue).Debug(err)
		return false
	}
	return boolValue
}
