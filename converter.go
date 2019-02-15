package bridges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"ngsi-bridge/ngsi"
)

func (h *HTTPBridge) decode(ctx *gin.Context) {
	ctx.Status(http.StatusBadRequest)

	buff, err := ioutil.ReadAll(ctx.Request.Body)
	buff = bytes.Replace(buff, []byte{'\\', '"'}, []byte{'"'}, -1)
	buff = bytes.Replace(buff, []byte("data="), []byte(""), 1)
	msg := make(map[string]interface{})
	err = json.Unmarshal(buff, &msg)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePrivate)
		ctx.Error(fmt.Errorf("field msg format can not be parsed to a map[string]interface")).SetType(gin.ErrorTypePublic)
		ctx.Abort()
		return
	}

	key := ctx.Param("key")
	if key == "" {
		keyI, ok := ctx.Get("key")
		if !ok {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("no schema found for %s", key)).SetType(gin.ErrorTypePublic)
			return
		}
		key = keyI.(string)
	}

	sch, ok := h.mapper[key]
	if !ok {
		ctx.Error(fmt.Errorf("no schema found for %s", key)).SetType(gin.ErrorTypePublic)
		ctx.Abort()
		return
	}

	if field := sch.Data.Field; field != "" {
		if msg, err = dataField(msg, field); err != nil {
			ctx.Error(err).SetType(gin.ErrorTypePublic)
			ctx.Abort()
			return
		}
	}
	msg = replaceField(msg, sch.Replace)

	id, err := findID(msg)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		ctx.Abort()
		return
	}

	attrs := make(map[string]ngsi.Attribute)
	for k, t := range sch.Attrs {
		if v, ok := msg[k]; ok {
			attrs[k] = ngsi.Attribute{
				AttrPair: ngsi.AttrPair{
					Value: v,
					Type:  t,
				},
			}
		} else {
			h.ctx.Warnf("no type for %s", k)
		}
	}
	attrs["timestamp"] = ngsi.Attribute{
		AttrPair: ngsi.AttrPair{
			Type:  "time",
			Value: time.Now().UTC(),
		},
	}
	ctx.Set("element", ngsi.Entity{
		Type:       "WaterTank",
		Id:         id,
		Attributes: attrs,
	})
}

func (h *HTTPBridge) Encode(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func findID(msg map[string]interface{}) (string, error) {
	iid, ok := msg["id"]
	if !ok {
		return "", fmt.Errorf("could not find id field")
	}
	id, ok := iid.(string)
	if !ok {
		return "", fmt.Errorf("id field could not be converted to string")
	}
	return id, nil
}

func replaceField(msg map[string]interface{}, keys map[string]string) map[string]interface{} {
	for key, rKey := range keys {
		if val, ok := msg[rKey]; ok {
			msg[key] = val
			delete(msg, rKey)
		}
	}
	return msg
}

func dataField(msg map[string]interface{}, field string) (map[string]interface{}, error) {
	data := msg
	split := strings.Split(field, ".")
	for _, str := range split {
		tmp, ok := data[str]
		if !ok {
			return nil, fmt.Errorf("could not find field %s", field)
		}
		switch tmp.(type) {
		case string:
			return nil, json.Unmarshal([]byte(tmp.(string)), &msg)
		case map[string]interface{}:
			data = tmp.(map[string]interface{})
		default:
			return nil, fmt.Errorf("unsupported field type %T", tmp)
		}
	}
	delete(msg, split[0])
	for key, val := range msg {
		data[key] = val
	}
	return data, nil
}
