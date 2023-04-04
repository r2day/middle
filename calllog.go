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
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	SYS_LOGIN_LOG           = "sys_login_log"
	defaultCallLogColl      = "default_call_log"
	defaultOperationLogColl = "default_operation_log"
	ignoreGET               = "GET"
)

var (
	customCallLogColl      = os.Getenv("CUSTOME_CALL_LOG_COLL")
	customOperationLogColl = os.Getenv("CUSTOME_OPERATION_LOG_COLL")
)

type CallLogData struct {
	// 创建时（用户上传的数据为空，所以默认可以不传该值)
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// 商户号
	MerchantId string `json:"-" bson:"merchant_id"`
	// 创建者
	UserId string `json:"user_id" bson:"user_id"`
	// 创建时间
	CreatedAt string `json:"created_at" bson:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at" bson:"updated_at"`
	// 更新人
	UpdatedUserId string `json:"updated_user_id" bson:"updated_user_id"`
	// 状态
	Status bool `json:"status"`
	// 客户IP
	ClientIP string `json:"client_ip" bson:"client_ip"`
	// 远程IP
	RemoteIP string `json:"remote_ip"  bson:"remote_ip"`
	// 路径
	FullPath string `json:"full_path"  bson:"full_path"`
	// 请求方法/操作
	Method string `json:"method"  bson:"method"`
	// 相应代码
	RespCode int `json:"resp_code"  bson:"resp_code"`
	// 目标
	TargetId string `json:"target_id"  bson:"target_id"`
}

// LoginRequest Binding from JSON
type LoginRequest struct {
	// 商户号
	MerchantId string `form:"merchant_id" json:"merchant_id" xml:"merchant_id"  binding:"required"`
	// Phone 手机号
	Phone string `form:"phone" json:"phone" xml:"phone"  binding:"required"`
	// Password 密码
	Password string `form:"password" json:"password" xml:"password"`
	// Type 用户类型
	Type string `form:"type" json:"type" xml:"type"`
}

// LoginLogMiddleware 登陆日志
func LoginLogMiddleware(db *mongo.Database) gin.HandlerFunc {

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

		if c.Request.Method == ignoreGET {
			logCtx.Debug("it is get method, we don't record it on database")
			c.Next()
			return
		}

		// 声明表
		coll := db.Collection(SYS_LOGIN_LOG)

		isSimpleSign, err := strconv.ParseBool(os.Getenv("IS_SIMPLE_SIGN"))
		if err != nil {
			log.Fatal(err)
		}

		if isSimpleSign {
			var jsonInstance body.SimpleSignInRequest
			if err := c.ShouldBindBodyWith(&jsonInstance, binding.JSON); err != nil {
				// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "login params no right"})
				logCtx.Error(err)
				c.Next()
				return
			}

			newOne := &CallLogData{}
			// 基本查询条件
			// newOne.MerchantId = jsonInstance.MerchantId
			newOne.ID = primitive.NewObjectID()
			newOne.UserId = jsonInstance.Phone

			// 插入身份信息
			createdAt := rtime.FomratTimeAsReader(time.Now().Unix())
			whoChange := c.GetHeader("AccountId")
			newOne.UserId = whoChange
			newOne.UpdatedUserId = whoChange
			newOne.CreatedAt = createdAt
			newOne.UpdatedAt = createdAt
			newOne.ClientIP = clientIP
			newOne.RemoteIP = remoteIP
			newOne.FullPath = fullPath
			newOne.RespCode = respCode

			// 写入数据库
			// 插入记录
			_, err := coll.InsertOne(c.Request.Context(), newOne)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "failed to insert one", "error": err.Error()})
				return
			}
		} else {
			var jsonInstance body.SignInRequest
			if err := c.ShouldBindBodyWith(&jsonInstance, binding.JSON); err != nil {
				logCtx.Error(err)
				c.Next()
				return
			}

			newOne := &CallLogData{}
			// 基本查询条件
			newOne.MerchantId = jsonInstance.MerchantId
			newOne.ID = primitive.NewObjectID()
			newOne.UserId = jsonInstance.Phone

			// 插入身份信息
			createdAt := rtime.FomratTimeAsReader(time.Now().Unix())
			whoChange := c.GetHeader("AccountId")
			newOne.UserId = whoChange
			newOne.UpdatedUserId = whoChange
			newOne.CreatedAt = createdAt
			newOne.UpdatedAt = createdAt
			newOne.ClientIP = clientIP
			newOne.RemoteIP = remoteIP
			newOne.FullPath = fullPath
			newOne.RespCode = respCode

			// 写入数据库
			// 插入记录
			_, err := coll.InsertOne(c.Request.Context(), newOne)
			if err != nil {
				logCtx.Error(err)
				c.Next()
				return
			}

		}
	}
}

// CallLogMiddleware 调用日志
func CallLogMiddleware(db *mongo.Database) gin.HandlerFunc {

	return func(c *gin.Context) {
		method := c.Request.Method
		if c.Request.Method == ignoreGET {
			fmt.Println("it is get method ,no data change so don't need to record it by default")
			c.Next()
			return
		}

		if customOperationLogColl == "" {
			customOperationLogColl = defaultOperationLogColl
		}

		// 声明表
		coll := db.Collection(customOperationLogColl)

		clientIP := c.ClientIP()
		remoteIP := c.RemoteIP()
		fullPath := c.FullPath()

		newOne := &CallLogData{}
		// 基本查询条件
		newOne.MerchantId = c.GetHeader("MerchantId")
		newOne.ID = primitive.NewObjectID()

		// 插入身份信息
		createdAt := rtime.FomratTimeAsReader(time.Now().Unix())
		whoChange := c.GetHeader("AccountId")
		newOne.UserId = whoChange
		newOne.UpdatedUserId = whoChange
		newOne.CreatedAt = createdAt
		newOne.UpdatedAt = createdAt
		newOne.ClientIP = clientIP
		newOne.RemoteIP = remoteIP
		newOne.FullPath = fullPath
		newOne.Method = method
		newOne.TargetId = c.Param("_id")

		// 写入数据库
		// 插入记录
		_, err := coll.InsertOne(c.Request.Context(), newOne)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to insert one", "error": err.Error()})
			return
		}
		c.Next()
	}
}
