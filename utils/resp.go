package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H struct {
	Code int
	Msg string
	Data interface{}
	Rows interface{}
	Total interface{}
}

func Resp(w http.ResponseWriter, code int, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H {
		Code : code,
		Data: data,
		Msg : msg,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	//将结果json传到响应
	w.Write(ret)
}

func RespList(w http.ResponseWriter, code int, data interface{}, total interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H {
		Code : code, //0
		Rows: data, //user切片
		Total: total,//切片长度
	}
	//将h转为json
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}


func RespFail(w http.ResponseWriter, msg string) {
	Resp(w, -1, nil, msg)
}


func RespOk(w http.ResponseWriter, data interface{}, msg string) {
	Resp(w, 0, data, msg)
}



func RespOKList(w http.ResponseWriter, data interface{}, total interface{}) {
	RespList(w,0,data,total)
}

