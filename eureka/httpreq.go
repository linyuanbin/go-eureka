package eureka

import (
	"crypto/tls"
	"net/http"

	"github.com/sirupsen/logrus"
)

func doHttpRequest(httpAction HttpAction) bool {
	req := httpAction.BuildHttpRequest()

	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	resp, err := DefaultTransport.RoundTrip(req)
	if resp != nil && resp.StatusCode > 299 {
		defer resp.Body.Close()
		logrus.Error("HTTP request failed with status: ", resp.StatusCode)
		return false
	} else if err != nil {
		logrus.Error("HTTP request failed: ", err.Error())
		return false
	} else {
		defer resp.Body.Close()
		return true
	}
}
