package global

// 数据请求类型
type RequestType int

const (
	AccessNumber RequestType = iota // 检查单
	UidEnc                          // 检查单ID
	InstanceKey                     // Instance_key
)

const (
	Interface_Type_Platform int = iota // 通过平台转发的下载模式
	Interfacce_Type_S3                 // 通过S3下载模式
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

// 文件状态
type FileStatus int

const (
	FileNotExist FileStatus = iota // 文件不存在
	FileExist                      // 文件存在
	FileFailed                     // 文件失败
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
