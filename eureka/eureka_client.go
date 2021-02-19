package eureka

import (
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var instanceId string
var interval = 3
var renewalIntervalInSecs = "30"
var durationInSecs = "90"

/**
服务提供方配置参考：
#服务过期时间配置,超过这个时间没有接收到心跳EurekaServer就会将这个实例剔除
#注意，EurekaServer一定要设置eureka.server.eviction-interval-timer-in-ms否则这个配置无效，这个配置一般为服务刷新时间配置的三倍
#默认90s
eureka.instance.lease-expiration-duration-in-seconds=90
#服务刷新时间配置，每隔这个时间会主动心跳一次
#默认30s
eureka.instance.lease-renewal-interval-in-seconds=10
*/
type Client struct {
	DiscoveryServerUrl    string
	Retry                 int64
	Heartbeat             int64
	BuildHeartbeat        bool
	renewalIntervalInSecs string
	durationInSecs        string
}

func NewEurekaClient(serverUrl string, retry int64, heartbeat int64) *Client {
	return &Client{DiscoveryServerUrl: serverUrl, Retry: retry, Heartbeat: heartbeat}
}

var regTpl = `{
  "instance": {
    "hostName":"${ipAddress}",
    "app":"${appName}",
    "ipAddr":"${ipAddress}",
    "vipAddress":"${appName}",
    "status":"UP",
    "port": {
      "$":${port},
      "@enabled": true
    },
    "securePort": {
      "$":${securePort},
      "@enabled": false
    },
    "homePageUrl" : "http://${ipAddress}:${port}/",
    "statusPageUrl": "http://${ipAddress}:${port}/actuator/info",
    "healthCheckUrl": "http://${ipAddress}:${port}/health",
    "dataCenterInfo" : {
      "@class":"com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
      "name": "MyOwn"
    },
    "metadata": {
      "instanceId" : "${appName}:${instanceId}"
    },
	"leaseInfo": {
      "renewalIntervalInSecs" : ${renewalIntervalInSecs},
      "durationInSecs" : ${durationInSecs}
	}
  }
}`

/**
@RenewalIntervalInSecs
服务刷新时间配置，每隔这个时间会主动心跳一次
默认30s
*/
func (client *Client) SetRenewalIntervalInSecs(renewalIntervalInSecs string) {
	client.renewalIntervalInSecs = renewalIntervalInSecs
}

/**
@DurationInSecs
注意，EurekaServer一定要设置eureka.server.eviction-interval-timer-in-ms否则这个配置无效，这个配置一般为服务刷新时间配置的三倍
#默认90s
*/
func (client *Client) SetDurationInSecs(durationInSecs string) {
	client.durationInSecs = durationInSecs
}

func (client *Client) isExist(appName string, port string) bool {
	service, er := client.GetServiceInstances(appName)
	if er != nil {
		Error("Registration attempt of " + appName + " failed!")
		Errorf("Registration attempt of " + appName + " failed!")
		return false
	}
	//避免重复注册
	if len(service) > 0 {
		status := false
		for _, s := range service {
			if getLocalIP() == s.HostName && port == strconv.Itoa(s.Port.Port) {
				status = true
				break
			}
		}
		if status {
			return true
		}
	}
	return false
}

//RegisterWithHost 注册服务
func (client *Client) RegisterWithHost(appName string, ip string, port string) {
	regTpl = strings.Replace(regTpl, "${ipAddress}", ip, -1)
	client.Register(appName, port)
}

//Register 注册服务(自动获取本地IP)
func (client *Client) Register(appName string, port string) {
	if client.isExist(appName, port) {
		//服务已经存在：服务中心正在，当前服务升级，需要重新建立心跳
		//避免非法重复注册
		if !client.BuildHeartbeat {
			handleSigterm(client, appName, instanceId)
			go client.startHeartbeat(client.DiscoveryServerUrl, appName, port)
			go client.reRegistration(appName, port)
		}
		return
	}
	instanceId = port
	tpl := string(regTpl)
	tpl = strings.Replace(tpl, "${ipAddress}", getLocalIP(), -1)
	tpl = strings.Replace(tpl, "${port}", port, -1)
	tpl = strings.Replace(tpl, "${securePort}", "443", -1)
	tpl = strings.Replace(tpl, "${instanceId}", instanceId, -1)
	tpl = strings.Replace(tpl, "${appName}", appName, -1)
	if len(client.durationInSecs) > 0 {
		tpl = strings.Replace(tpl, "${durationInSecs}", client.durationInSecs, -1)
	} else {
		tpl = strings.Replace(tpl, "${durationInSecs}", durationInSecs, -1)
	}
	if len(client.renewalIntervalInSecs) > 0 {
		tpl = strings.Replace(tpl, "${renewalIntervalInSecs}", client.renewalIntervalInSecs, -1)
	} else {
		tpl = strings.Replace(tpl, "${renewalIntervalInSecs}", renewalIntervalInSecs, -1)
	}
	// Register.
	registerAction := HttpAction{
		Url:         client.DiscoveryServerUrl + "apps/" + appName,
		Method:      "POST",
		ContentType: "application/json;charset=UTF-8",
		Body:        tpl,
	}
	var result bool
	var try int64 = 0
	for {
		result = doHttpRequest(registerAction)
		if result {
			client.BuildHeartbeat = true
			Info("Registration Success !")
			handleSigterm(client, appName, instanceId)
			go client.startHeartbeat(client.DiscoveryServerUrl, appName, port)
			go client.reRegistration(appName, port)
			break
		} else {
			Warn("Registration attempt of " + appName + " failed!")
			Warn("url: ", registerAction.Url)
			time.Sleep(time.Second * time.Duration(interval))
			try++
			if try >= client.Retry {
				break
			}
		}
	}
}

func (client *Client) register(appName string, port string) {
	instanceId = port
	tpl := string(regTpl)
	tpl = strings.Replace(tpl, "${ipAddress}", getLocalIP(), -1)
	tpl = strings.Replace(tpl, "${port}", port, -1)
	tpl = strings.Replace(tpl, "${securePort}", "443", -1)
	tpl = strings.Replace(tpl, "${instanceId}", instanceId, -1)
	tpl = strings.Replace(tpl, "${appName}", appName, -1)
	if len(client.durationInSecs) > 0 {
		tpl = strings.Replace(tpl, "${durationInSecs}", client.durationInSecs, -1)
	} else {
		tpl = strings.Replace(tpl, "${durationInSecs}", durationInSecs, -1)
	}
	if len(client.renewalIntervalInSecs) > 0 {
		tpl = strings.Replace(tpl, "${renewalIntervalInSecs}", client.renewalIntervalInSecs, -1)
	} else {
		tpl = strings.Replace(tpl, "${renewalIntervalInSecs}", renewalIntervalInSecs, -1)
	}
	// Register.
	registerAction := HttpAction{
		Url:         client.DiscoveryServerUrl + "apps/" + appName,
		Method:      "POST",
		ContentType: "application/json;charset=UTF-8",
		Body:        tpl,
	}
	var result bool
	var try int64 = 0
	for {
		result = doHttpRequest(registerAction)
		if result {
			Info("Registration Success !")
			handleSigterm(client, appName, instanceId)
			break
		} else {
			Warn("Registration attempt of " + appName + " failed!")
			Warn("url: ", registerAction.Url)
			time.Sleep(time.Second * time.Duration(interval))
			try++
			if try >= client.Retry {
				break
			}
		}
	}
}

//reRegistration 注册中心离线恢复，自动重新注册
func (client *Client) reRegistration(appName string, port string) {
	for {
		time.Sleep(time.Second * time.Duration(client.Heartbeat))
		if client.isExist(appName, port) {
			continue
		}
		client.register(appName, port)
	}
}

func (client *Client) Deregister(appName string, instanceId string) {
	Info("Trying to deregister application " + appName + "...")
	// Deregister
	deregisterAction := HttpAction{
		Url:         client.DiscoveryServerUrl + "apps/" + appName + "/" + getLocalIP() + ":" + appName + ":" + instanceId,
		ContentType: "application/json;charset=UTF-8",
		Method:      "DELETE",
	}
	a := doHttpRequest(deregisterAction)
	if !a {
		Error("Deregister ", appName, " failed")
	}
	Info("Deregistered application " + appName + ", exiting. Check Eureka...")
}

//GetServiceInstances 获取服务信息
func (client *Client) GetServiceInstances(appName string) (rest []Instance, err error) {
	var m ServiceResponse
	Info("Querying eureka for instances of " + appName + " at: " + client.DiscoveryServerUrl + "apps/" + appName)
	queryAction := HttpAction{
		Url:         client.DiscoveryServerUrl + "apps/" + appName,
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	Info("Doing queryAction using URL: " + queryAction.Url)
	bytes, err := queryAction.Do()
	if err != nil {
		Error("GetServiceInstances", err.Error())
		return
	} else {
		Info("Got instances response from Eureka:\n" + string(bytes))
		if len(bytes) == 0 {
			return
		}
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			Error("Problem parsing JSON response from Eureka: " + err.Error())
			return nil, err
		}
		rest = m.Application.Instance
	}
	return
}

//GetServiceUrl 获取服务地址
func (client *Client) GetServiceUrl(appName string) (url string) {
	sers, er := client.GetServiceInstances(appName)
	if er != nil {
		return
	}
	if len(sers) > 0 {
		url = "http://" + sers[0].HostName + ":" + strconv.Itoa(sers[0].Port.Port) + "/"
	}
	return
}

//GetServices 获取Eureka服务中心服务列表
func (client *Client) GetServices() (rest []Application, err error) {
	var m ApplicationsRootResponse
	Info("Querying eureka for services at: " + client.DiscoveryServerUrl + "apps")
	queryAction := HttpAction{
		Url:         client.DiscoveryServerUrl + "apps",
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	bytes, err := queryAction.Do()
	if err != nil {
		Error(err.Error(), err)
		return
	} else {
		if len(bytes) == 0 {
			return
		}
		err := json.Unmarshal(bytes, &m)
		if err != nil {
			Error("Unmarshal response from Eureka: " + err.Error())
			return nil, err
		}
		rest = m.Resp.Applications
	}
	return
}

func (client *Client) startHeartbeat(url string, appName string, port string) {
	for {
		time.Sleep(time.Second * time.Duration(client.Heartbeat))
		heartbeat(url, appName, port)
	}
}

func heartbeat(url string, appName string, port string) {
	heartbeatAction := HttpAction{
		//apps/monitor/10.95.58.86:monitor:18085?status=UP
		Url:         url + "apps/" + appName + "/" + getLocalIP() + ":" + appName + ":" + port + "?status=UP",
		Method:      "PUT",
		ContentType: "application/json;charset=UTF-8",
	}
	Info("Issuing heartbeat to " + heartbeatAction.Url)
	doHttpRequest(heartbeatAction)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("Unable to determine local IP address (non loopback). Exiting.")
}

func handleSigterm(client *Client, appName string, instanceId string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		client.Deregister(appName, instanceId)
		os.Exit(1)
	}()
}
