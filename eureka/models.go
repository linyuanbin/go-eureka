package eureka

// ServiceResponse 按服务名称查询注册服务响应
type ServiceResponse struct {
	Application Application `json:"application"`
}

//ApplicationsRootResponse 服务列表
type ApplicationsRootResponse struct {
	Resp ApplicationsResponse `json:"applications"`
}

type ApplicationsResponse struct {
	Version      string        `json:"versions__delta"`
	AppsHashcode string        `json:"versions__delta"`
	Applications []Application `json:"application"`
}

type Application struct {
	Name     string     `json:"name"`
	Instance []Instance `json:"instance"`
}

type Instance struct {
	HostName string `json:"hostName"`
	Port     Port   `json:"port"`
}

type Port struct {
	Port int `json:"$"`
}
