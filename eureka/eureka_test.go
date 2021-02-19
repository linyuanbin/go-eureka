package eureka

import (
	"gopkg.in/go-playground/assert.v1"
	"testing"
)

var client *Client

func init() {
	client = NewEurekaClient("http://localhost:8761", 5, 30)
	client.SetDurationInSecs("90")
	client.SetRenewalIntervalInSecs("30")
	//init log object
	InitLog(log)
}

func TestEurekaClient_GetServices(t *testing.T) {
	//client.RegisterWithHost("TEST_SERVICE_NAME","127.0.0.1","8080")

	client.Register("TEST_SERVICE_NAME", "8080")

	_, err := client.GetServices()
	assert.Equal(t, nil, err)
}

func TestEurekaClient_GetServiceInstances(t *testing.T) {
	_, er := client.GetServiceInstances("TEST_SERVICE_NAME")
	assert.Equal(t, nil, er)
}
