package service

import (
	"fmt"
	"io"
	"math/rand"
	"mychat/internal/common"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Image(ctx *gin.Context) {
	w := ctx.Writer
	req := ctx.Request

	srcFile, head, err := req.FormFile("file")
	if err != nil {
		common.RespFail(w, err.Error())
		return
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
		common.RespFail(w, err.Error())
		return
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		common.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	common.RespOK(w, url, "发送成功")
}
