package router

import (
	"dev/converge_alert_mail/converge_alert_mail_server/common"
	"dev/converge_alert_mail/converge_alert_mail_server/mongo"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

var config common.Config = common.GetConf()
var logger *zap.Logger = common.InitLogger()

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	var user common.UserCredentials
	var response common.Token
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		logger.Error(fmt.Sprintf("[LoginHandler] Error in request, %v", err))
		return
	}

	if strings.ToLower(user.SecretId) != config.SecretId {
		if user.SecretKey != config.SecretKey {
			w.WriteHeader(http.StatusForbidden)
			logger.Error(fmt.Sprintf("[LoginHandler] Invalid credentials, %v", err))
			return
		}
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(fmt.Sprintf("[LoginHandler] Error extracting the key, %v", err))
	}

	tokenString, err := token.SignedString([]byte(common.SecretKey))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(fmt.Sprintf("[LoginHandler] Error while signing the token, %v", err))
		response.Status = false
		response.Token = ""
		response.Msg = fmt.Sprintf("Error while signing the token, %v.", err)
	} else {
		response.Status = true
		response.Token = tokenString
		response.Msg = "get token success"
	}

	JsonResponse(response, w)

}

func ValidateTokenMiddleware(ss string) (err error) {
	token, err := jwt.Parse(ss, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(common.SecretKey), nil
	})
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	} else {
		fmt.Println(err)
		return err
	}
}

func DeviceTimeOutHandler(w http.ResponseWriter, r *http.Request) {
	var data common.DeviceTimeOutDoc
	var response common.Response
	token := r.FormValue("token")
	response.Status = false
	if err := ValidateTokenMiddleware(token); err != nil {

		response.Msg = "token 验证不通过"
	}
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		logger.Error(fmt.Sprintf("[DeviceTimeOutHandler] Error in request, %v", err))
		response.Msg = "内容解析失败"
	}
	collection, _ := mongo.MongoClient(config.Mongodb.Url, config.Mongodb.Db, config.Mongodb.Table)
	err = mongo.MgoInsert(collection, data.Project, data.Module, data, logger)
	if err != nil {
		response.Msg = "数据插入失败"
	} else {
		response.Status = true
	}
	JsonResponse(response, w)
}

func DeviceCycleHandler(w http.ResponseWriter, r *http.Request) {
	var data common.DeviceCycleDoc
	var response common.Response
	token := r.FormValue("token")
	response.Status = false
	if err := ValidateTokenMiddleware(token); err != nil {

		response.Msg = "token 验证不通过"
	}
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		logger.Error(fmt.Sprintf("[DeviceCycleHandler] Error in request, %v", err))
		response.Msg = "内容解析失败"
	}
	collection, _ := mongo.MongoClient(config.Mongodb.Url, config.Mongodb.Db, config.Mongodb.Table)
	err = mongo.MgoInsert(collection, data.Project, data.Module, data, logger)
	if err != nil {
		response.Msg = "数据插入失败"
	} else {
		response.Status = true
	}
	JsonResponse(response, w)
}

func JsonResponse(response interface{}, w http.ResponseWriter) {

	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
