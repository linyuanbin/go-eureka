### Go程序实现Spring Cloud Eureka注册中心服务注册

#### EurekaClient使用

* 服务注册
```
func main() {
    //注册中心地址
  eurekaClient := NewEurekaClient("http://localhost:8761/eureka/")

   //如果注册中心关闭了自身保护模式，可以自定义设置服务列表刷新时间（一般不建议关闭自身保护模式）
	eurekaClient.SetRenewalIntervalInSecs("30")
    eurekaClient.SetDurationInSecs("90")

    //日志对象初始化： logger 传自己的日志对象
    eureka.InitLog(logger)

    //服务名称 、端口 (默认获取本地IP)
    eurekaClient.Register("monitor", "8080")

    //也可以指定IP注册
    eurekaClient.RegisterWithHost("monitor","127.0.0.1", "8080")
    
}
```

```
客户端提供  /health 接口，服务中心定时采集状态（只是状态采集）
engine.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, `{"status":"UP"}`)
	})
```


* 服务发现
```
func main() {
    //注册中心地址
  eurekaClient := NewEurekaClient("http://localhost:8761/eureka/")
    //获取服务列表
  serviceList, err := eurekaClient.GetServices()
    //根据服务名称获取服务信息
  service, er := eurekaClient.GetServiceInstances("TEST_SERVICE_NAME")

    //获取服务地址 (例如：http://ip:port/)
    url := eurekaClient.GetServiceUrl("TEST_SERVICE_NAME")
}  
```

#### 兼容

* 向服务中心同步心跳
* 服务中心重启，服务自动注册
* 当前服务升级，自动重新建立心跳机制(注册中心不升级，当前服务升级，不需要重新注册，但是需要重新建立心跳机制)
* 注册到注册中心的statusPageUrl (http://host:port/actuator/info)
* 服务刷新时间支持自定义
* 日志打印支持自定义
