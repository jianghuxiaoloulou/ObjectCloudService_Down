package object

import (
	"WowjoyProject/ObjectCloudService_Down/global"
	"WowjoyProject/ObjectCloudService_Down/internal/model"
	"WowjoyProject/ObjectCloudService_Down/pkg/errcode"
	"WowjoyProject/ObjectCloudService_Down/pkg/general"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

//var token string

// 封装对象相关操作
type Object struct {
	Key        int64             // 目标key
	FileKey    string            // 文件key
	FilePath   string            // 文件路径
	ActionType global.ActionMode // 操作类型
	FileType   global.FileModel  // 文件类型
	Count      int               // 文件执行次数
}

func NewObject(data global.ObjectData) *Object {
	return &Object{
		Key:        data.InstanceKey,
		FileKey:    data.FileKey,
		FilePath:   data.FilePath,
		ActionType: data.ActionType,
		FileType:   data.FileType,
		Count:      data.Count,
	}
}

// 上传对象[POST]
func (obj *Object) UploadObject() {
	global.Logger.Info("开始上传对象：", *obj)
	code := UploadFile(obj)
	if code == "00000" {
		//上传成功更新数据库
		global.Logger.Info("数据上传成功", obj.Key)
		model.UpdateUplaod(obj.Key, obj.FileType, true)
	} else {
		global.Logger.Info("数据上传失败", obj.Key)
		// 上传失败时先补偿操作，补偿操作失败后才更新数据库
		if !ReDo(obj) {
			global.Logger.Info("数据补偿失败", obj.Key)
			// 上传失败更新数据库
			model.UpdateUplaod(obj.Key, obj.FileType, false)
		}
	}
}

// 下载对象[GET]
func (obj *Object) DownObject() {
	global.Logger.Info("开始下载对象：", *obj)
	flag := DownFile(obj)
	if flag {
		global.Logger.Info("下载成功：" + obj.FilePath)
		model.UpdateDown(obj.Key, obj.FileType, true)
	} else {
		// 下载失败时先补偿操作，补偿操作失败后才更新数据库
		if !ReDo(obj) {
			global.Logger.Info("数据补偿失败", obj.Key)
			// 下载失败更新数据库
			model.UpdateDown(obj.Key, obj.FileType, false)
		}
	}
}

// UploadFile 上传文件
func UploadFile(obj *Object) string {
	global.Logger.Debug("开始执行文件上传")
	url := global.ObjectSetting.OBJECT_POST_Upload
	url += "//"
	url += global.ObjectSetting.OBJECT_ResId
	url += "//"
	url += obj.FileKey
	global.Logger.Debug("操作的URL: ", url)
	file, err := os.Open(obj.FilePath)
	if err != nil {
		return errcode.File_OpenError.Msg()
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("file", obj.FilePath)
	if err != nil {
		global.Logger.Error("CreateFormFile err :", err, file)
		return errcode.Http_HeadError.Msg()
	}
	_, err = io.Copy(formFile, file)
	if err != nil {
		return errcode.File_CopyError.Msg()
	}

	writer.Close()
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		global.Logger.Error("NewRequest err: ", err, url)
		return errcode.Http_RequestError.Msg()
	}
	// 设置AK
	request.Header.Set("accessKey", global.ObjectSetting.OBJECT_AK)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Connection", "close")
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(request)
	if err != nil {
		global.Logger.Error("Do Request got err: ", err)
		return errcode.Http_RequestError.Msg()
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errcode.Http_RespError.Msg()
	}
	global.Logger.Info("resp.Body: ", string(content))
	var result = make(map[string]interface{})
	err = json.Unmarshal(content, &result)
	if err != nil {
		global.Logger.Error("resp.Body: ", "错误")
		return errcode.Http_RespError.Msg()
	}
	// 解析json
	if vCode, ok := result["code"]; ok {
		resultcode := vCode.(string)
		global.Logger.Info("resultcode: ", resultcode)
		return resultcode
	}
	return ""
}

// 补偿操作
func ReDo(obj *Object) bool {
	global.Logger.Info("开始补偿操作：", obj.Key)
	if obj.Count < global.ObjectSetting.OBJECT_Count {
		obj.Count += 1
		data := global.ObjectData{
			InstanceKey: obj.Key,
			FileKey:     obj.FileKey,
			FilePath:    obj.FilePath,
			ActionType:  obj.ActionType,
			FileType:    obj.FileType,
			Count:       obj.Count,
		}
		global.ObjectDataChan <- data
		return true
	}
	return false
}

// DownFile 下载文件
func DownFile(obj *Object) bool {
	url := global.ObjectSetting.OBJECT_GET_Download
	url += "//"
	url += global.ObjectSetting.OBJECT_ResId
	url += "//"
	url += obj.FileKey
	global.Logger.Debug("操作的URL: ", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("文件下载失败", err, obj.Key)
		return false
	}
	// 设置AK
	req.Header.Set("accessKey", global.ObjectSetting.OBJECT_AK)
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error(err)
		return false
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	if code != 200 {
		global.Logger.Error(resp.StatusCode, "下载失败："+obj.FilePath)
		return false
	}
	len, _ := strconv.ParseInt(resp.Header.Get("Content-size"), 10, 64)
	global.Logger.Info("获取的文件长度：", len)
	general.CheckPath(obj.FilePath)
	file, _ := os.Create(obj.FilePath)
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		global.Logger.Error("下载失败：文件拷贝失败：" + obj.FilePath)
		file.Close()
		os.Remove(obj.FilePath)
		return false
	} else {
		size := general.GetFileSize(obj.FilePath)
		global.Logger.Info("下载文件获取的长度：", size)
		if size != len {
			global.Logger.Error("下载失败：保存的文件大小错误：" + obj.FilePath)
			file.Close()
			os.Remove(obj.FilePath)
			return false
		}
	}
	return true
}
