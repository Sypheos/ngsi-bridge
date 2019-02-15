package ngsi

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"

	"github.com/TheThingsNetwork/go-utils/log"
)

func TestSubscribeEntityType(t *testing.T) {

	a := assertions.New(t)
	ctx := log.Get()
	resp, err := SubscribeEntityType(ctx, "https://projects.thethings.industries/broker", "http://ea75b2ad.ngrok.io", "WaterTank", []string{"temp1"})
	a.So(err, assertions.ShouldBeNil)
	fmt.Println(resp)
}
