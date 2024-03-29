package routers

import (
	v1 "WowjoyProject/ObjectCloudService_Down/internal/routers/api/v1"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 注册中间件
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiv1 := r.Group("/api/v1")
	{
		// 通过检查号上传数据
		apiv1.POST("/Object/Upload/AccessNumber/:AccessNumber", v1.ByAccessNunUpload)
		// 通过uid_enc上传数据
		apiv1.POST("/Object/Upload/UidEnc/:UidEnc", v1.ByUidEncUpload)
		// 通过instanceKey上传数据
		apiv1.POST("/Object/Upload/InstanceKey/:InstanceKey", v1.ByInstanceKeyUpload)
		// 通过检查号下载数据
		apiv1.GET("/Object/Down/AccessNumber/:AccessNumber", v1.ByAccessNumDownData)
		// 通过uid_enc下载数据
		apiv1.GET("/Object/Down/UidEnc/:UidEnc", v1.ByUidEncDownData)
		// 通过instance下载数据
		apiv1.GET("/Object/Down/InstanceKey/:InstanceKey", v1.ByInstanceKeyDownData)

		// test
		apiv1.POST("/SaveFile", v1.ByAccessNunUpload)
	}
	return r
}
