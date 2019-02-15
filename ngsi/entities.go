package ngsi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/TheThingsNetwork/go-utils/log"
)

const (
	entities   = "/v2/entities"
	entity     = entities + "/%s"
	entityAttr = entity + "/attrs"
)

type Entity struct {
	Id         string               `json:"id,omitempty"`
	IdPattern  string               `json:"idPattern,omitempty"`
	Type       string               `json:"type"`
	Attributes map[string]Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	AttrPair
	Metadata map[string]AttrPair `json:"metadata,omitempty"`
}

type AttrPair struct {
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

func (e Entity) MarshalJSON() ([]byte, error) {
	type entity_ Entity
	e_ := entity_{
		Type:      e.Type,
		Id:        e.Id,
		IdPattern: e.IdPattern,
	}
	buff, err := json.Marshal(e_)
	if err != nil {
		return nil, err
	}
	if e.Attributes != nil {
		buff = buff[:len(buff)-1]
		buff_, err := json.Marshal(e.Attributes)
		if err != nil {
			return nil, err
		}
		buff_[0] = ','
		buff = append(buff, buff_...)
	}
	return buff, nil
}

var re = regexp.MustCompile(`(?m)"id":"\w*",("idPattern":"\w*",)?"type":"\w*"[,}]`)

func (e *Entity) UnmarshalJSON(b []byte) error {
	loc := re.FindIndex(b)
	b[loc[1]-1] = '}'
	type _entity Entity
	var ent _entity
	err := json.Unmarshal(b[:loc[1]], &ent)
	if err != nil {
		return err
	}
	e.Id = ent.Id
	e.Type = ent.Type
	e.IdPattern = ent.IdPattern
	if len(b) > loc[1] {
		b[loc[1]-1] = '{'
		err = json.Unmarshal(b[loc[1]-1:], &e.Attributes)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterEntity add a new entity in the broker with the attributes present in the entity struct
func RegisterEntity(ctx log.Interface, brokerURL string, entity *Entity) error {
	ctx.Infof("Registering... entityId=%s", entity.Id)
	if _, err := request(fmt.Sprintf(brokerURL+entities), "POST", entity); err != nil {
		return fmt.Errorf("failed to register entity: %s", err.Error())
	}
	ctx.Infof("Registered entityId=%s", entity.Id)
	return nil
}

// PushAttributes update an entity attributes. it use the POST method in update mode so if an attributes is missing it
// will be created and the same field won't appear twice. If the entity doesn't exist it will attempt to create it.
func PushAttributes(ctx log.Interface, brokerURL string, entity *Entity) error {
	ctx.Infof("Push data entityId=%s", entity.Id)
	if _, err := request(fmt.Sprintf(brokerURL+entityAttr, entity.Id), "POST", entity.Attributes); err != nil {
		if strings.Index(err.Error(), "code=404") != -1 {
			ctx.Infof("Entity not registered %v", err)
			RegisterEntity(ctx, brokerURL, entity)
		} else {
			return fmt.Errorf("failed to push attributes: %s", err.Error())
		}
	}
	ctx.Infof("Pushed data entityId=%s", entity.Id)
	return nil
}
