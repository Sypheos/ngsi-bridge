package ngsi

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
)

func TestEntityJSON(t *testing.T) {
	testTable := [][]byte{
		[]byte(`{"id":"hello","idPattern":"testPattern","type":"test"}`),
		[]byte(`{"id":"hello","idPattern":"testPattern","type":"test","testValue":{"value":"test","type":"testType"}}`),
		[]byte(`{"id":"45001d000551353437353039","type":"WaterTank","precipitation":{"type":"L.mm","value":"0.00","metadata":{}},"release":{"type":"bool","value":"0.000000","metadata":{}},"stateOfCharge":{"type":"percentage","value":"81.2","metadata":{}},"temp1":{"type":"celcius","value":"6.00","metadata":{}},"temp2":{"type":"celcius","value":"0.00","metadata":{}},"timestamp":{"type":"time","value":"2019-01-16T16:42:31.717492461Z","metadata":{}},"waterlevel":{"type":"meter","value":"1.6","metadata":{}},"windspeed":{"type":"m.s","value":"0.0","metadata":{}}}`),
	}
	a := assertions.New(t)
	ent := Entity{
		Id:        "hello",
		Type:      "test",
		IdPattern: "testPattern",
	}
	buff, err := json.Marshal(ent)
	a.So(err, assertions.ShouldBeNil)
	a.So(buff, assertions.ShouldResemble, testTable[0])
	tmp := Entity{}
	err = json.Unmarshal(testTable[0], &tmp)
	a.So(err, assertions.ShouldBeNil)
	a.So(tmp, assertions.ShouldResemble, ent)

	ent.Attributes = map[string]Attribute{
		"testValue": {
			AttrPair: AttrPair{
				Value: "test",
				Type:  "testType",
			},
			Metadata: nil,
		},
	}
	buff, err = json.Marshal(ent)
	a.So(err, assertions.ShouldBeNil)
	a.So(buff, assertions.ShouldResemble, testTable[1])
	err = json.Unmarshal(testTable[1], &tmp)
	a.So(err, assertions.ShouldBeNil)
	a.So(tmp, assertions.ShouldResemble, ent)
}
