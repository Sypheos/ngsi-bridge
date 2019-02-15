package bridges

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ngsi-bridge/ngsi"
)

func decode(msg map[string]interface{}, sch *Schema) (*ngsi.Entity, error) {
	var err error
	if field := sch.Data.Field; field != "" {
		if msg, err = dataField(msg, field); err != nil {
			return nil, err
		}
	}
	msg = replaceField(msg, sch.Replace)

	id, err := findID(msg)
	if err != nil {
		return nil, err
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
		}
	}
	attrs["timestamp"] = ngsi.Attribute{
		AttrPair: ngsi.AttrPair{
			Type:  "time",
			Value: time.Now().UTC(),
		},
	}
	return &ngsi.Entity{
		Type:       "WaterTank",
		Id:         id,
		Attributes: attrs,
	}, nil
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
