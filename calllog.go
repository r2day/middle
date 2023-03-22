package middle

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	rtime "github.com/r2day/base/time"
	"github.com/gin-gonic/gin/binding"
)

const (
	defaultCallLogColl = "default_call_log"
)

var (
	customCallLogColl = os.Getenv("CUSTOME_CALL_LOG_COLL")
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

// CallLogMiddleware 调用日志
func CallLogMiddleware(db * mongo.Database) gin.HandlerFunc {

	return func(c *gin.Context) {

	if (customCallLogColl == "") {
		customCallLogColl = defaultCallLogColl
	}

	// 声明表
	coll := db.Collection(customCallLogColl)

	var jsonInstance LoginRequest
	if err := c.ShouldBindBodyWith(&jsonInstance, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "request params no right"})
		return
	}

	clientIP := c.ClientIP()
	remoteIP := c.RemoteIP()
	fullPath := c.FullPath()


	newOne := &CallLogData{}
	// 基本查询条件
	newOne.MerchantId = jsonInstance.MerchantId
	newOne.ID = primitive.NewObjectID()
	newOne.UserId = jsonInstance.Phone

	// 插入身份信息
	createdAt := rtime.FomratTimeAsReader(time.Now().Unix())
	whoChange := "frank"
	newOne.UserId = whoChange
	newOne.UpdatedUserId = whoChange
	newOne.CreatedAt = createdAt
	newOne.UpdatedAt = createdAt
	newOne.ClientIP = clientIP
	newOne.RemoteIP = remoteIP
	newOne.FullPath = fullPath

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