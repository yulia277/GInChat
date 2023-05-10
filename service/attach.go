package service

import (
	"IMProject/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	_"net/http"
	"os"
	"strings"
	"time"
)

func Upload(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	suffix := ".png"
	ofileName := head.Filename
	tem := strings.Split(ofileName, ".")
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1]
	}

	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	dstFile, err := os.Create("./asset/upload/" + fileName)

	if err != nil {
		utils.RespFail(w, err.Error())
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		utils.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	utils.RespOk(w,url ,"发送图片成功")
}
