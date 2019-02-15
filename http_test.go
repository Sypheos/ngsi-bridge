package bridges

import (
	"testing"

	"github.com/TheThingsIndustries/integration-messaging/ngsi10/ngsi10"
)

func TestHttpBridge_Particle(t *testing.T) {
	bridge := NewHttpBridge(8080)
	bridge.Prepare(nil, map[string]*ngsi10.Schema{}, "localhost:1026/ngsi10/updateContext", "POST")
}
