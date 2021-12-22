package main

import (
	"WowjoyProject/ObjectCloudService_Down/global"
	"WowjoyProject/ObjectCloudService_Down/internal/routers"
	"WowjoyProject/ObjectCloudService_Down/pkg/object"
	"WowjoyProject/ObjectCloudService_Down/pkg/workpattern"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @title 存储策略下载服务
// @version 1.0.0.1
// @description 存储策略下载服务
// @termsOfService https://github.com/jianghuxiaoloulou/ObjectCloudService_Down.git
func main() {
	global.Logger.Info("***开始运行存储策略下载服务***")
	global.ObjectDataChan = make(chan global.ObjectData)
	// 注册工作池，传入任务
	// 参数1 初始化worker(工人)设置最大线程数
	wokerPool := workpattern.NewWorkerPool(global.GeneralSetting.MaxThreads)
	// 有任务就去做，没有就阻塞，任务做不过来也阻塞
	wokerPool.Run()
	// 处理任务
	go func() {
		for {
			select {
			case data := <-global.ObjectDataChan:
				sc := &Dosomething{key: data}
				wokerPool.JobQueue <- sc
			}
		}
	}()
	web()
}

type Dosomething struct {
	key global.ObjectData
}

func (d *Dosomething) Do() {
	global.Logger.Info("正在处理的数据是：", d.key)
	//处理封装对象
	obj := object.NewObject(d.key)
	switch d.key.ActionType {
	case global.UPLOAD:
		// 数据上传
		obj.UploadObject()
	case global.DOWNLOAD:
		// 数据下载
		obj.DownObject()
	}
}

func web() {
	gin.SetMode(global.ServerSetting.RunMode)
	router := routers.NewRouter()

	ser := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeout,
		WriteTimeout:   global.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	ser.ListenAndServe()
}
