// Copyright © 2018 The Things Industries B.V.

package ngsi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func request(brokerURL, method string, elem interface{}) ([]byte, error) {
	req, err := prepareRequest(brokerURL, method, elem)
	if err != nil {
		return nil, fmt.Errorf("ngsi prepare request failed: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	buff, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		if err != nil {
			return nil, err
		}
		return buff, fmt.Errorf("request failed requestURL=%v code=%v body=%v", req.URL, resp.StatusCode, string(buff))
	}
	return buff, err
}

// prepareRequest prepare a request to be sent to the IoT broker. JSON content-type/
func prepareRequest(uri string, method string, message interface{}) (*http.Request, error) {
	var body []byte
	var err error
	body, err = json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("message not marchallized: %v", err)
	}
	req, err := http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, err
}
