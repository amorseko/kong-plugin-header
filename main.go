package main

import (
	"encoding/json"
	"fmt"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

type Config struct {
    // Nama objek yang akan diubah, misalnya "Header" atau "Data"
    TargetField string `json:"target_field"`
    // Nama kunci yang akan ditambahkan ke target_field
    Key string `json:"key"`
    // Nama header atau query parameter yang akan digunakan untuk mendapatkan nilai
    ValueSource string `json:"value_source"`
}

func New() interface{} {
    return &Config{}
}

func (conf *Config) Access(kong *pdk.PDK) {
    var value string
    var err error

    switch conf.ValueSource {
    case "header":
        headers, err := kong.Request.GetHeaders(-1)
        if err != nil {
            kong.Log.Err(fmt.Sprintf("failed to get headers: %v", err))
            return
        }
        values := headers[conf.Key]
        if len(values) > 0 {
            value = values[0] // Ambil nilai pertama dari header
        }
    case "query":
		query, err := kong.Request.GetQueryArg(conf.Key)
		if err != nil {
            kong.Log.Err(fmt.Sprintf("failed to get query params: %v", err))
            return
        }
        value = query
    default:
        kong.Log.Err(fmt.Sprintf("unknown value_source: %v", conf.ValueSource))
        return
    }

    if value == "" {
        value = "default_value"
    }

    body, err := kong.Request.GetRawBody()
    if err != nil {
        kong.Log.Err(fmt.Sprintf("failed to get body: %v", err))
        return
    }

    var bodyMap map[string]interface{}
    err = json.Unmarshal(body, &bodyMap)
    if err != nil {
        kong.Log.Err(fmt.Sprintf("failed to unmarshal body: %v", err))
        return
    }

    if targetField, exists := bodyMap[conf.TargetField].(map[string]interface{}); exists {
        targetField[conf.Key] = value
    } else {
        bodyMap[conf.TargetField] = map[string]interface{}{
            conf.Key: value,
        }
    }

    modifiedBody, err := json.Marshal(bodyMap)
    if err != nil {
        kong.Log.Err(fmt.Sprintf("failed to marshal modified body: %v", err))
        return
    }

    err = kong.ServiceRequest.SetRawBody(string(modifiedBody))
    if err != nil {
        kong.Log.Err(fmt.Sprintf("failed to set modified body: %v", err))
        return
    }
}

func main() {
    server.StartServer(New, "0.1", 1)
}