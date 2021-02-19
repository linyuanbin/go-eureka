package eureka

import (
	"testing"
)

func TestLog(t *testing.T) {
	Info("hello")
	Infof("hello:%s-%s", "123", "456")
}
