package middle

import (
	"encoding/base64"
	"encoding/json"
)

type LoginInfo struct {
	// 命名空间
	// 可是商户号
	Namespace string `json:"namespace"`
	// 用户名
	// 可以是手机号/用户名
	User string `json:"user"`
	// Avatar 用户头像
	Avatar string `json:"avatar"`
}

// 登陆信息
func DumpLoginInfo(namespace string, user string, avatar string) string{
	// step 01 转换为json
	loginInfo := LoginInfo{
		Namespace: namespace,
		User: user,
		Avatar: avatar,
	}
	payload, err := json.Marshal(loginInfo)
	if err != nil {
		panic(err)
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(payload))
	return sEnc
}

// 登陆信息
func LoadLoginInfo(payload string) * LoginInfo {
	// step 01 转换为bytes
	sDec, err  := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		panic(err)
	}
	loginInfo := &LoginInfo{}
	err = json.Unmarshal(sDec, loginInfo)
	if err != nil {
		panic(err)
	}
	return loginInfo
}