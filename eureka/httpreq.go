package eureka

import (
	"crypto/tls"
	"net/http"
)

func doHttpRequest(httpAction HttpAction) bool {
	req := httpAction.BuildHttpRequest()

	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	resp, err := DefaultTransport.RoundTrip(req)
	if resp != nil && resp.StatusCode > 299 {
		defer resp.Body.Close()
		Infof("HTTP request failed with status: %d", resp.StatusCode)
		return false
	} else if err != nil {
		Infof("HTTP request failed: %s", err.Error())
		return false
	} else {
		defer resp.Body.Close()
		return true
	}
}
