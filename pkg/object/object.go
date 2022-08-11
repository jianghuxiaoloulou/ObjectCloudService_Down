package object

import (
	"WowjoyProject/ObjectCloudService_Down/global"
	"WowjoyProject/ObjectCloudService_Down/internal/model"
	"WowjoyProject/ObjectCloudService_Down/pkg/errcode"
	"WowjoyProject/ObjectCloudService_Down/pkg/general"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
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
	var code string
	// 增加上传模式，是通过平台上传还是临时地址上传
	if global.ObjectSetting.OBJECT_Interface_Type == global.Interfacce_Type_S3 {
		global.Logger.Info("***通过S3接口上传数据***")
		code = S3UploadFile(obj)
	} else {
		global.Logger.Info("***通过平台接口转发上传数据***")
		code = UploadFile(obj)
	}
	if code == "00000" {
		//上传成功更新数据库
		global.Logger.Info("数据上传成功", obj.Key)
		model.UpdateUplaod(obj.Key, obj.FileType, obj.FileKey, true)
	} else {
		global.Logger.Info("数据上传失败", obj.Key)
		// 上传失败时先补偿操作，补偿操作失败后才更新数据库
		if !ReDo(obj) {
			global.Logger.Info("数据补偿失败", obj.Key)
			// 上传失败更新数据库
			model.UpdateUplaod(obj.Key, obj.FileType, obj.FileKey, false)
		}
	}
}

// 下载对象[GET]
func (obj *Object) DownObject() {
	global.Logger.Info("开始下载对象：", *obj)
	var flag bool
	// 增加上传模式，是通过平台上传还是临时地址上传
	if global.ObjectSetting.OBJECT_Interface_Type == global.Interfacce_Type_S3 {
		global.Logger.Info("***通过S3接口下载数据***")
		flag = DownS3File(obj)
	} else {
		global.Logger.Info("***通过平台接口转发下载数据***")
		flag = DownFile(obj)
	}

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

// 通过S3临时地址下载文件
func DownS3File(obj *Object) bool {
	// 1.获取S3临时地址
	global.Logger.Debug("开始获取临时地址")
	url := global.ObjectSetting.OBJECT_Temp_GET_Down
	url += "//"
	url += global.ObjectSetting.OBJECT_ResId
	url += "//"
	url += obj.FileKey
	global.Logger.Debug("操作的URL: ", url)
	err, s3url := GetS3URL(url)
	if err != nil {
		global.Logger.Error("获取S3临时上传地址错误", err)
		return false
	}
	// 2.通过临时地址下载
	global.Logger.Debug("开始通过临时地址下载：", s3url)
	return Down_S3(s3url, obj)
}

// 获取S3临时地址
func GetS3URL(url string) (error, string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("http.NewRequest err", err)
		return err, ""
	}
	// 设置AK
	req.Header.Set("accessKey", global.ObjectSetting.OBJECT_AK)
	// 设置参数
	q := req.URL.Query()
	q.Add("expireTime", "60000")
	req.URL.RawQuery = q.Encode()
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error("client.do err", err)
		return err, ""
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != 200 {
		global.Logger.Error("获取临时地址失败:", resp.StatusCode)
		return errcode.Http_RespError, ""
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("ioutil.ReadAll err: ", err)
		return errcode.Http_RespError, ""
	}
	global.Logger.Info("resp.Body: ", string(content))
	var result = make(map[string]interface{})
	err = json.Unmarshal(content, &result)
	if err != nil {
		global.Logger.Error("resp.Body: ", "错误")
		return errcode.Http_RespError, ""
	}
	// 解析json
	if UrlData, ok := result["data"]; ok {
		resultUrl := UrlData.(string)
		global.Logger.Info("resultUrl: ", resultUrl)
		return nil, resultUrl
	}
	return errcode.Http_RespError, ""
}

func Down_S3(url string, obj *Object) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("文件下载失败", err, obj.Key)
		return false
	}
	// 设置AK
	// req.Header.Set("accessKey", global.ObjectSetting.OBJECT_AK)
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
	general.CheckPath(obj.FilePath)
	file, _ := os.Create(obj.FilePath)
	defer file.Close()

	// buff := make([]byte, 1024)
	// for {
	// 	n, err := resp.Body.Read(buff)
	// 	if err != nil && err != io.EOF {
	// 		global.Logger.Error("resp.Body.Read err ", err)
	// 		return false
	// 	}
	// 	if err == io.EOF {
	// 		// file.WriteAt()
	// 		break
	// 	}
	// 	_, err = file.Write(buff[:n])
	// 	if err != nil {
	// 		global.Logger.Error("file write err: ", err)
	// 		file.Close()
	// 		os.Remove(obj.FilePath)
	// 		global.Logger.Error("下载失败,写文件失败" + obj.FilePath)
	// 		return false
	// 	}
	// }

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		global.Logger.Error("下载失败：文件拷贝失败：" + obj.FilePath)
		file.Close()
		os.Remove(obj.FilePath)
		return false
	}
	return true
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
		global.Logger.Error("Open File err :", err)
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
		global.Logger.Error("File Copy err :", err)
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
	connectTimeout := 60 * time.Second
	readWriteTimeout := 60 * time.Second
	transport := http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		Dial:              TimeoutDialer(connectTimeout, readWriteTimeout),
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(request)
	global.Logger.Info("开始发起http client.Do: ", obj.Key)
	if err != nil {
		global.Logger.Error("Do Request got err: ", err)
		return errcode.Http_RequestError.Msg()
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("ioutil.ReadAll err: ", err)
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

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// S3接口直接上传数据
func S3UploadFile(obj *Object) string {
	// 1.获取临时上传地址
	global.Logger.Debug("开始获取临时地址")
	url := global.ObjectSetting.OBJECT_Temp_GET_Upload
	url += "//"
	url += global.ObjectSetting.OBJECT_ResId
	url += "//"
	url += obj.FileKey
	global.Logger.Debug("操作的URL: ", url)
	err, s3url := UploadGetS3URL(url)
	if err != nil {
		global.Logger.Error("获取S3临时上传地址错误", err)
		return err.Error()
	}
	// 2.通过临时上传地址上传数据
	global.Logger.Debug("开始通过临时地址上传：", s3url)
	return Upload_S3(s3url, obj)
}

// 获取S3临时上传地址
func UploadGetS3URL(url string) (error, string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("http.NewRequest err", err)
		return err, ""
	}
	// 设置AK
	req.Header.Set("accessKey", global.ObjectSetting.OBJECT_AK)
	req.Header.Set("Connection", "close")
	connectTimeout := 20 * time.Second
	readWriteTimeout := 20 * time.Second

	// 设置参数
	q := req.URL.Query()
	q.Add("expireTime", "60000")
	req.URL.RawQuery = q.Encode()
	transport := http.Transport{
		DisableKeepAlives: true,
		Dial:              TimeoutDialer(connectTimeout, readWriteTimeout),
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error("client.do err", err)
		return err, ""
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != 200 {
		global.Logger.Error("获取临时地址失败:", resp.StatusCode)
		return errcode.Http_RespError, ""
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("ioutil.ReadAll err: ", err)
		return errcode.Http_RespError, ""
	}
	global.Logger.Info("resp.Body: ", string(content))
	var result = make(map[string]interface{})
	err = json.Unmarshal(content, &result)
	if err != nil {
		global.Logger.Error("resp.Body: ", "错误")
		return errcode.Http_RespError, ""
	}
	// 解析json
	if UrlData, ok := result["data"]; ok {
		resultUrl := UrlData.(string)
		global.Logger.Info("resultUrl: ", resultUrl)
		return nil, resultUrl
	}
	return errcode.Http_RespError, ""
}

// S3上传数据
func Upload_S3(url string, obj *Object) string {
	fileSize := general.GetFileSize(obj.FilePath)

	file, err := os.Open(obj.FilePath)
	if err != nil {
		global.Logger.Error("Open File err :", err)
		return errcode.File_OpenError.Msg()
	}
	defer file.Close()
	body := &bytes.Buffer{}
	if fileSize >= (int64(global.ObjectSetting.File_Fragment_Size << 20)) {
		// 大文件分块读取
		buff := make([]byte, 1024)
		for {
			n, err := file.Read(buff)
			// 控制条件，根据实际调整
			if err != nil && err != io.EOF {
				global.Logger.Error(err)
				return ""
			}
			if n == 0 {
				break
			}
			body.Write(buff[:n])
		}
	} else {
		// 小文件直接读取
		body.ReadFrom(file)
	}

	global.Logger.Info("http.NewRequest 开始请求上传文件", obj.Key)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		global.Logger.Error("http.NewRequest err", err)
		return err.Error()
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Connection", "close")

	transport := http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error("client.do err", err)
		return ""
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	global.Logger.Debug("S3上传数据 resp.StatusCode:", resp.StatusCode)
	if code == 200 {
		return "00000"
	}
	return ""
}
