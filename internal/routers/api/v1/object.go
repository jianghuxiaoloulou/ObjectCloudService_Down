package v1

import (
	"WowjoyProject/ObjectCloudService_Down/global"
	"WowjoyProject/ObjectCloudService_Down/internal/model"
	"WowjoyProject/ObjectCloudService_Down/pkg/app"
	"WowjoyProject/ObjectCloudService_Down/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// 通过检查号上传
func ByAccessNunUpload(c *gin.Context) {
	id := c.Param("AccessNumber")
	global.Logger.Info("需要上传的检查号是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取上传任务：
		model.GetRequestData(id, global.AccessNumber, global.UPLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}

// 通过UidEnc上传
func ByUidEncUpload(c *gin.Context) {
	id := c.Param("UidEnc")
	global.Logger.Info("需要下载的UidEnc是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取下载任务：
		model.GetRequestData(id, global.UidEnc, global.UPLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}

// 通过InstanceKey上传数据
func ByInstanceKeyUpload(c *gin.Context) {
	id := c.Param("InstanceKey")
	global.Logger.Info("需要下载的InstanceKey是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取下载任务：
		model.GetRequestData(id, global.InstanceKey, global.UPLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}

// 通过检查号下载
func ByAccessNumDownData(c *gin.Context) {
	id := c.Param("AccessNumber")
	global.Logger.Info("需要下载的检查号是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取下载任务：
		model.GetRequestData(id, global.AccessNumber, global.DOWNLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}

// 通过UidEnc下载数据
func ByUidEncDownData(c *gin.Context) {
	id := c.Param("UidEnc")
	global.Logger.Info("需要下载的UidEnc是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取下载任务：
		model.GetRequestData(id, global.UidEnc, global.DOWNLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}

// 通过InstanceKey下载数据
func ByInstanceKeyDownData(c *gin.Context) {
	id := c.Param("InstanceKey")
	global.Logger.Info("需要下载的InstanceKey是：", id)
	if id != "" {
		// 成功：
		app.NewResponse(c).ToResponse(nil)
		// 获取下载任务：
		model.GetRequestData(id, global.InstanceKey, global.DOWNLOAD)
	} else {
		// 失败：
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	}
}
