package restapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func writeInternalServerErrorResponse(ctx *gin.Context, err error) {
	glog.Errorf("http internal server error - %s: err: %v", ctx.Request.RequestURI, err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"err": "internal server error"})
}

func writeNotFoundResponse(ctx *gin.Context, msg string) {
	glog.Infof("http not found response: %s", msg)
	ctx.JSON(http.StatusNotFound, gin.H{"msg": msg})
}

func writeLockResponse(ctx *gin.Context, msg string) {
	glog.Infof("http locked response: %s", msg)
	ctx.JSON(http.StatusLocked, gin.H{"msg": msg})
}

func writeBadRequestResponse(ctx *gin.Context, msg string) {
	glog.Infof("bad request: %s", msg)
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": msg})
}
