package middle

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	rtime "github.com/r2day/base/time"
	"github.com/r2day/body"
	"github.com/r2day/collections/clog"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	defaultCallLogColl      = "default_call_log"
	defaultOperationLogColl = "default_operation_log"
)

var (
	customCallLogColl      = os.Getenv("CUSTOME_CALL_LOG_COLL")
	customOperationLogColl = os.Getenv("CUSTOME_OPERATION_LOG_COLL")
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

// CallLogMiddleware 调用日志
func CallLogMiddleware(db *mongo.Database) gin.HandlerFunc {

	return func(c *gin.Context) {
		method := c.Request.Method
		if c.Request.Method == http.MethodGet {
			fmt.Println("it is get method ,no data change so don't need to record it by default")
			c.Next()
			return
		}

		if customOperationLogColl == "" {
			customOperationLogColl = defaultOperationLogColl
		}

		clientIP := c.ClientIP()
		remoteIP := c.RemoteIP()
		fullPath := c.FullPath()

		// 声明表
		m := &clog.Model{}
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

		// 写入数据库
		// 插入记录
		_, err := m.Create(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to insert one", "error": err.Error()})
			return
		}
		c.Next()
	}
}
