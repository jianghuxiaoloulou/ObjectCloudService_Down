# 项目
# ****云对象存储服务****

# 项目描述
<!-- 实现云数据下载 -->

# 设计依据
<!-- 根据检查号或者UidEnc从数据库查询数据实现下载 -->

# 目录结构
<!-- 
configs：配置文件。
docs：文档集合。
global：全局变量。
internal：内部模块。
model：数据库相关操作。
routers：路由相关逻辑处理。
pkg：项目相关的模块包。
storage：项目生成的临时文件。
scripts：各类构建，安装，分析等操作的脚本。
-->
# 路由
<!-- 在 RESTful API 中 HTTP 方法对应的行为动作分别如下：

GET：读取/检索动作。
POST：新增/新建动作。
PUT：更新动作，用于更新一个完整的资源，要求为幂等。
PATCH：更新动作，用于更新某一个资源的一个组成部分，也就是只需要更新该资源的某一项，就应该使用 PATCH 而不是 PUT，可以不幂等。
DELETE：删除动作。 -->

# 对象管理
功能(一次检查的所有数据)	       HTTP方法       路径
上传对象	                      POST	        /object/:id
删除指定对象	                   DELETE	     /object/:id
更新指定对象	                   PUT	         /object/:id
获取对象                           GET	         /object/:id

# 公共组件
错误码标准化
配置管理
数据库连接
日志写入
响应处理

# 文件配置文件读取：go get -u github.com/spf13/viper
Viper 是适用于GO 应用程序的完整配置解决方案

# 日志：go get -u gopkg.in/natefinch/lumberjack.v2
它的核心功能是将日志写入滚动文件中，该库支持设置所允许单日志文件的最大占用空间、最大生存周期、允许保留的最多旧文件数，
如果出现超出设置项的情况，就会对日志文件进行滚动处理。

# 生成接口文档
Swagger 相关的工具集会根据 OpenAPI 规范去生成各式各类的与接口相关联的内容，
常见的流程是编写注解 =》调用生成库-》生成标准描述文件 =》生成/导入到对应的 Swagger 工具
$ go get -u github.com/swaggo/swag/cmd/swag@v1.6.5
$ go get -u github.com/swaggo/gin-swagger@v1.2.0 
$ go get -u github.com/swaggo/files
$ go get -u github.com/alecthomas/template

@Summary	摘要
@Produce	API 可以产生的 MIME 类型的列表，MIME 类型你可以简单的理解为响应类型，例如：json、xml、html 等等
@Param	参数格式，从左到右分别为：参数名、入参类型、数据类型、是否必填、注释
@Success	响应成功，从左到右分别为：状态码、参数类型、数据类型、注释
@Failure	响应失败，从左到右分别为：状态码、参数类型、数据类型、注释
@Router	路由，从左到右分别为：路由地址，HTTP 方法

swag init

http://127.0.0.1:8000/swagger/index.html

# 接口做参数校验
在本项目中我们将使用开源项目 go-playground/validator 作为我们的本项目的基础库，
它是一个基于标签来对结构体和字段进行值验证的一个验证器。
标签	含义
required	必填
gt	大于
gte	大于等于
lt	小于
lte	小于等于
min	最小值
max	最大值
oneof	参数集内的其中之一
len	长度要求与 len 给定的一致

# 国际化处理
中间件
# 邮件报警处理
go get -u gopkg.in/gomail.v2
# 接口限流控制
go get -u github.com/juju/ratelimit@v1.0.1
# 统一超时控制


## 第二次相同项目提交文件到github
# git add README.md
# git commit -m "first commit"
# git push -u origin master


# 修改记录
# 2021-12-22 搭建存储策略下载服务
# 2021-12-28 增加本地文件状态，存在的不下载
# 2022-01-04 修改下载文件远端文件KEY值从数据库获取
# 2022/01/06 修改了表结构，增加了file_remote表，上传数据的所有内容从file_remote表中获取
# 2022/01/12 下载时增加判断是否存在云端（只有云端上有的数据才下载） 
# 2022/01/12 增加remote_key 的转移斜杠
# 2022/05/30 增加数据库读写分离