package main

import (
	"encoding/json"
	"fmt"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

type Config struct {
    TargetField    string `json:"target_field"`
    KeyValuePairs  []struct {
        Key          string `json:"key"`
        ValueSource  string `json:"value_source"`
    } `json:"key_value_pairs"`
}

func New() interface{} {
    return &Config{}
}

func (conf *Config) Access(kong *pdk.PDK) {
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

    if conf.TargetField != "" {
        targetField, exists := bodyMap[conf.TargetField].(map[string]interface{})
        if !exists {
            targetField = make(map[string]interface{})
            bodyMap[conf.TargetField] = targetField
        }

        for _, kv := range conf.KeyValuePairs {
            var value interface{}
            switch kv.ValueSource {
            case "header":
                headers, err := kong.Request.GetHeaders(-1)
                if err != nil {
                    kong.Log.Err(fmt.Sprintf("failed to get headers: %v", err))
                    continue
                }
                values, exists := headers[kv.Key]
                if exists && len(values) > 0 {
                    value = values[0]
                }
            case "query":
                queryArgs, err := kong.Request.GetQuery(-1)
                if err != nil {
                    kong.Log.Err(fmt.Sprintf("failed to get query: %v", err))
                    continue
                }
                values, exists := queryArgs[kv.Key]
                if exists && len(values) > 0 {
                    value = values[0]
                }
            default:
                kong.Log.Err(fmt.Sprintf("unknown value_source: %v", kv.ValueSource))
                continue
            }

            if value == nil {
                value = "default_value"
            }
            targetField[kv.Key] = value
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