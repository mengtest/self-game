package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gitlab.oss.snaplingo.com/server-team/common/utils/async"
	"net/http"
	"self_game/config"
	"self_game/constants/gameCode"
	"self_game/service"
	"self_game/utils"
	"self_game/utils/vo"
	"strings"
)

// 用户注册
func RegisterUserHandler(c *gin.Context) {
	retData := &vo.Data{}
	defer SendResponse(c, retData)

	var (
		err         error
		requestBody service.UserRegisterReq
		uid         string
	)

	if err = ParsePostBody(c, &requestBody); err != nil {
		logger.Errorf("uname=%v,err=%v", requestBody.UserName, err.Error())
		retData.Code = gameCode.RequestParamsError
		return
	}
	if err = service.CheckUserRegisterParams(requestBody); err != nil {
		retData.Code = gameCode.RequestParamsError
		logger.Error(err)
		return
	}
	// 检查该用户是否存在
	err = service.CheckUserIsExist(requestBody.UserName)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			retData.Code = gameCode.UserNameAlreadyExist
			logger.Error(err)
			return
		}
	}

	// save user to db
	if uid, err = service.InsertUserToDB(requestBody); err != nil {
		retData.Code = gameCode.RequestParamsError
		return
	}

	// 更改玩家的ip,国家和城市
	go async.Do(func() {
		ip := c.ClientIP()
		err = service.UpdateUserCountryAndCity(uid, ip)
		if err != nil {
			logger.Errorf("ip=%v,err=%v", ip, err)
			return
		}
	})

	logger.Infof("userRegister:%v", requestBody)
	retData.Data = map[string]interface{}{
		"uid":           uid,
		"register_time": utils.GetTimeZoneTime(config.Config.Cfg.TimeZone).Format("2006-01-02 15:04:05"),
	}
	retData.Code = gameCode.RequestSuccess
	return
}

func HandlerSignatureHandler(c *gin.Context) {
	var (
		signature string
		echostr   string
		timestamp int
		nonce     int
		err       error
	)

	if signature, echostr, timestamp, nonce, err = service.GetSignatrueParams(c); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(signature, echostr, timestamp, nonce)

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})

}

func GetUserNameHandler(c *gin.Context) {
	retData := &vo.Data{}
	defer SendResponse(c, retData)
	fmt.Println("hello")
	var (
		uid string
	)
	uid = c.Param("uid")
	if strings.TrimSpace(uid) == "12345" {
		retData.Code = -101
		retData.Data = "param error"
		return
	}

	retData.Data = map[string]interface{}{
		"name": config.Config.Cfg.Port,
		"env":  config.Config.Env.ENV,
	}
	retData.Code = 1
	return
}

func ConsulHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func PostUserNameHandler(c *gin.Context) {
	retData := vo.NewData()
	defer SendResponse(c, retData)

	var (
		res  service.PostUserRequest
		resp service.PostUserResponse
		err  error
		uid  string
	)

	if err = ParsePostBody(c, &res); err != nil {
		retData.Data = err.Error()
		retData.Code = -101
		logger.Error("param error")
		return
	}
	fmt.Println("name=", res.Name, "english_score=", res.EnglishScore)

	if strings.TrimSpace(res.Name) == "" {
		retData.Data = "params error"
		retData.Code = -101
		logger.Error("param error")
		return
	}

	resp.UID = uid
	resp.Name = res.Name
	resp.ChineseScore = res.EnglishScore + 2
	resp.EnglishScore = res.EnglishScore
	retData.Code = 1
	retData.Data = map[string]interface{}{
		"user_info": resp,
	}
	logger.Infof("userName=%v,score=%v,responseBody=%v", res.Name, res.EnglishScore, resp)
	return
}
