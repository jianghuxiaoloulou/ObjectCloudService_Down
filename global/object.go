package global

// 数据请求类型
type RequestType int

const (
	AccessNumber RequestType = iota // 检查单
	UidEnc                          // 检查单ID
)

// 数据处理模式
type ActionMode int

const (
	UPLOAD   ActionMode = iota // 上传
	DOWNLOAD                   // 下载
	DELETE                     // 删除
)

// 文件模式
type FileModel int

const (
	DCM FileModel = iota // DCM 文件
	JPG                  // JPG 文件
)

const (
	PublicCloud  int = iota // 共有云
	PrivateCloud            // 私有云
)

type ObjectData struct {
	InstanceKey int64
	FileKey     string     // 文件key
	FilePath    string     // 文件路径
	ActionType  ActionMode // 操作类型
	FileType    FileModel  // 文件类型
	Count       int        // 执行次数
}

var (
	ObjectDataChan chan ObjectData
)
