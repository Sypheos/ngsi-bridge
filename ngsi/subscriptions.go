package ngsi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TheThingsNetwork/go-utils/log"
)

const (
	subscription = "/v2/subscriptions"
)

type Subscription struct {
	Description  string          `json:"description"`
	Subject      subSubject      `json:"subject"`
	Notification subNotification `json:"notification"`
	Expires      string          `json:"expires,omitempty"`
	Throttling   int             `json:"throttling"`
}

type subSubject struct {
	Entities  []Entity     `json:"entities"`
	Condition subCondition `json:"condition"`
}

type subCondition struct {
	Attrs      []string          `json:"attrs"`
	Expression map[string]string `json:"expression,omitempty"`
}

type subNotification struct {
	Http struct {
		Url string `json:"url"`
	} `json:"http"`
	Attrs       []string `json:"attrs,omitempty"`
	ExceptAttrs []string `json:"exceptAttrs,omitempty"`
}

type Notification struct {
	Data  []Entity `json:"data"`
	SubID string   `json:"subscriptionID"`
}

type SubList []string

func SubscriptionServer(ctx log.Interface, port string, ch chan *Notification) {
	engine := gin.New()
	engine.Use(
		notificationLog(ctx),
		gin.ErrorLoggerT(gin.ErrorTypePublic),
		//checkHeaders
	)
	engine.POST("/", func(context *gin.Context) {
		n := &Notification{}
		err := context.BindJSON(n)
		if err != nil {
			ctx.WithError(err).Warn("Could not bind json to type")
			context.Status(http.StatusBadRequest)
			return
		}
		ch <- n
	})
	ctx.Infof("Starting subscriptions server...")
	engine.Run(":" + port)
	ctx.Info("subscriptions server started.")
}

func notificationLog(ctx log.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		fields := log.Fields{
			"Method":        c.Request.Method,
			"Path":          c.Request.URL.Path,
			"Query":         c.Request.URL.Query(),
			"Host":          c.Request.Host,
			"Id":            c.GetHeader("X-Request-ID"),
			"RemoteAddress": c.ClientIP(),
			"Duration":      end.Sub(start),
			"Status":        c.Writer.Status(),
		}
		if err := c.Errors.ByType(gin.ErrorTypePrivate).Last(); err != nil {
			ctx = ctx.WithError(err)
		}
		ctx.WithFields(fields).Info("Inbound notification")
	}
}

func SubscribeEntityType(ctx log.Interface, brokerURL, downURL, entityType string, attrs []string) (string, error) {
	ctx.Infof("Subscribing to entity type %s on attribute %v", entityType, attrs)
	sub := Subscription{
		Description: fmt.Sprintf("Subscription for %s on attrs %v", entityType, attrs),
		Subject: subSubject{
			Entities: []Entity{
				{
					IdPattern: ".*",
					Type:      entityType,
				},
			},
			Condition: subCondition{
				Attrs:      attrs,
				Expression: nil,
			},
		},
		Notification: subNotification{
			Http: struct {
				Url string `json:"url"`
			}{Url: downURL},
			ExceptAttrs: []string{"theattributeyoushoulnotuse"},
		},
		Expires: time.Now().AddDate(5, 0, 0).Format(time.RFC3339),
	}

	var body []byte
	var err error
	if body, err = request(brokerURL+subscription, "POST", sub); err != nil {
		return "", err
	}
	fmt.Println(string(body))
	return "", nil
}
