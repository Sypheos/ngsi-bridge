package bridges

import (
	"fmt"
	"net/http"
	"ngsi-bridge/ngsi"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/gin-gonic/gin"
)

// HTTPBridge define the http endpoint. It use the http framework to handle the HTTP request and a fiware.Agent to
// contact Fiware IoT broker
type HTTPBridge struct {
	ctx       log.Interface
	port      int
	engine    *gin.Engine
	mapper    map[string]Schema
	brokerURL string
	method    string
}

type Schema struct {
	Type    string
	Attrs   map[string]string
	Replace map[string]string
	ID      string
	Data    struct {
		Field  string
		Format string
	}
}

func NewHttpBridge(port int) *HTTPBridge {
	return &HTTPBridge{
		port: port,
	}
}

// Prepare the HTTP server. This a non blocking call
func (h *HTTPBridge) Prepare(ctx log.Interface, mapper map[string]Schema, brokerURL, method string) (err error) {
	h.ctx = ctx.WithField("endpoint", "HTTP")
	h.ctx.Info("Building bridge...")
	h.mapper = mapper
	h.brokerURL = brokerURL
	h.method = method
	h.engine = gin.New()
	h.engine.Use(
		Logger(h.ctx),
		gin.ErrorLoggerT(gin.ErrorTypePublic),
		//checkHeaders
	)
	h.engine.GET("/", h.Schemas)
	key := "/:key"
	h.engine.POST("/", func(context *gin.Context) {
		context.Set("key", "particle")
	}, h.decode, h.push)
	h.engine.POST(key, h.decode, h.push)
	h.engine.POST(key+"/register", h.decode, h.register)
	h.engine.GET(key, h.Encode)
	h.ctx.Info("Bridge built.")
	return nil
}

func (h *HTTPBridge) Open() error {
	return h.engine.Run(fmt.Sprintf(":%d", h.port))
}

func (h *HTTPBridge) Close() error {
	return nil
}

func (h *HTTPBridge) push(context *gin.Context) {
	ent, err := retrieveEntity(context)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err = ngsi.PushAttributes(h.ctx, h.brokerURL, ent); err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
	}
	context.Status(http.StatusOK)
}

func (h *HTTPBridge) register(context *gin.Context) {
	ent, err := retrieveEntity(context)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err = ngsi.RegisterEntity(h.ctx, h.brokerURL, ent); err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
	}
	context.Status(http.StatusOK)
}

func retrieveEntity(context *gin.Context) (*ngsi.Entity, error) {
	it, ok := context.Get("element")
	if !ok {
		return nil, fmt.Errorf("no element pushed")
	}
	elem, ok := it.(ngsi.Entity)
	if !ok {
		return nil, fmt.Errorf("cannot convert to ngsi element")
	}
	return &elem, nil
}

func (h *HTTPBridge) Schemas(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.mapper)
}
