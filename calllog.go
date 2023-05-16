package middle

import (
	"github.com/r2day/collections/auth/operation"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	rtime "github.com/r2day/base/time"
	"github.com/r2day/body"
	clog "github.com/r2day/collections/auth/log"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	defaultCallLogColl      = "default_call_log"
	defaultOperationLogColl = "default_operation_log"
)

var (
	customCallLogColl      = os.Getenv("CUSTOM_CALL_LOG_COLL")
	customOperationLogColl = os.Getenv("CUSTOM_OPERATION_LOG_COLL")
)

var (
	method2operation = map[string]string{
		"GET":    "查看",
		"POST":   "创建",
		"PUT":    "更新",
		"DELETE": "删除",
	}
)

// LoginLogMiddleware 登陆日志
func LoginLogMiddleware(db *mongo.Database, skipViewLog bool) gin.HandlerFunc {

	return func(c *gin.Context) {

		// 先执行登陆操作
		c.Next()
		// 获取用户登陆信息
		clientIP := c.ClientIP()
		remoteIP := c.RemoteIP()
		fullPath := c.FullPath()
		respCode := c.Writer.Status()

		logCtx := log.WithField("client_id", clientIP).
			WithField("remote_ip", remoteIP).
			WithField("full_path", fullPath).
			WithField("resp_status", respCode)

		if c.Request.Method == http.MethodGet && skipViewLog {
			logCtx.Debug("it is get method, we don't record it on database")
			return
		}

		// 声明表
		m := &clog.Model{}

		isSimpleSign, err := strconv.ParseBool(os.Getenv("IS_SIMPLE_SIGN"))
		if err != nil {
			logCtx.Error(err)
			return
		}

		if isSimpleSign {
			var jsonInstance body.SimpleSignInRequest
			if err := c.ShouldBindBodyWith(&jsonInstance, binding.JSON); err != nil {
				// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "login params no right"})
				logCtx.Error(err)
				return
			}
			// 基本查询条件
			// m.MerchantId = jsonInstance.MerchantId
			m.ID = primitive.NewObjectID()
			m.AccountID = c.GetHeader("AccountId")
			m.MerchantID = c.GetHeader("MerchantId")
			// 插入身份信息
			createdAt := rtime.FomratTimeAsReader(time.Now().Unix())

			m.CreatedAt = createdAt
			m.UpdatedAt = createdAt
			m.ClientIP = clientIP
			m.RemoteIP = remoteIP
			m.FullPath = fullPath
			m.RespCode = respCode

			// 写入数据库
			// 插入记录
			_, err := m.Create(c.Request.Context())
			if err != nil {
				logCtx.Error(err)
				return
			}
		} else {
			var jsonInstance body.SignInRequest
			if err := c.ShouldBindBodyWith(&jsonInstance, binding.JSON); err != nil {
				logCtx.Error(err)
				return
			}

			m.ID = primitive.NewObjectID()
			m.AccountID = c.GetHeader("AccountId")
			m.MerchantID = c.GetHeader("MerchantId")
			// 插入身份信息
			createdAt := rtime.FomratTimeAsReader(time.Now().Unix())

			m.CreatedAt = createdAt
			m.UpdatedAt = createdAt
			m.ClientIP = clientIP
			m.RemoteIP = remoteIP
			m.FullPath = fullPath
			m.RespCode = respCode

			// 写入数据库
			// 插入记录
			_, err := m.Create(c.Request.Context())
			if err != nil {
				logCtx.Error(err)
				return
			}

		}
	}
}

// OperationMiddleware 操作日志
func OperationMiddleware(db *mongo.Database, skipViewLog bool) gin.HandlerFunc {

	return func(c *gin.Context) {
		// 获取用户登陆信息
		clientIP := c.ClientIP()
		remoteIP := c.RemoteIP()
		fullPath := c.FullPath()
		method := c.Request.Method

		logCtx := log.WithField("client_id", clientIP).
			WithField("remote_ip", remoteIP).
			WithField("full_path", fullPath).
			WithField("method", method)

		if c.Request.Method == http.MethodGet && skipViewLog {
			// 如果是开启查看模式跳过，那么直接返回
			logCtx.Debug("it is get method, we don't record it on database")
			c.Next()
			return
		}

		// 声明表
		m := &operation.Model{}
		// 基本查询条件
		m.MerchantID = c.GetHeader("MerchantId")
		m.ID = primitive.NewObjectID()

		// 插入身份信息
		createdAt := rtime.FomratTimeAsReader(time.Now().Unix())

		m.CreatedAt = createdAt
		m.UpdatedAt = createdAt
		m.ClientIP = clientIP
		m.RemoteIP = remoteIP
		m.FullPath = fullPath
		m.Method = method
		m.TargetID = c.Param("_id")
		m.Operation = method2operation[method]
		// 通过path查找接口名称
		keyPrefix := AccessKeyPrefix + "_" + m.AccountID
		keyPath2Name := keyPrefix + "_" + "path2name"
		val, err := db.RDB.HGet(c.Request.Context(), keyPath2Name, fullPath).Result()
		if err != nil {
			// 可以忽略该日志
			// 一般情况下仅角色匹配到path即可访问
			// 其他角色大部分会走该逻辑，因此将日志类别定义为debug
			log.WithField("message", "call db.RDB.HGet failed").
				WithField("val", val).
				WithField("fullPath", fullPath).
				WithField("keyPath2Name", keyPath2Name).
				Debug(err)
			// 无法查找到路径对应的名称
			//c.Next()
			//return
		}
		m.Name = val
		// 写入数据库
		// 插入记录
		_, err = m.Create(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to insert one", "error": err.Error()})
			return
		}
		c.Next()
	}
}
