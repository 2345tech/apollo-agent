package apollo

// 定义请求polling 或 watching 的数据结构
// 定义获取到的配置数据的存储结构

type ConfigData struct {
	AppId          string
	Namespace      string
	Cluster        string
	Configurations map[string]string
}
