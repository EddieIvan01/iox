package operate

import "testing"

func TestProxyLocal(t *testing.T) {
	ProxyLocal(":9999", false)
}
