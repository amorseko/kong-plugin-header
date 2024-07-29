package main

import (
	"encoding/json"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	server.StartServer(New, "0.1", 1)
}

type Config struct{}

func New() interface{} {
	return &Config{}
}

func (conf *Config) Access(kong *pdk.PDK) {
	L := lua.NewState()
	defer L.Close()

	headers, err := kong.Request.GetHeaders(-1)
	if err != nil {
		kong.Log.Err("Failed to get headers: ", err.Error())
		return
	}

	payorCode := ""
	if val, ok := headers["payor_code"]; ok {
		if len(val) > 0 {
			payorCode = val[0]
		}
	}

	L.SetGlobal("payor_code", lua.LString(payorCode))
	if err := L.DoString(`
		payor_code = payor_code
		print("Payor code:", payor_code)
	`); err != nil {
		kong.Log.Err("Failed to execute Lua code: ", err.Error())
	}

	if payorCode == "" {
		payorCode = "default_payor_code"
	}

	rawBody, err := kong.Request.GetRawBody()
	if err != nil {
		kong.Log.Err("Failed to get raw body: ", err.Error())
		return
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(rawBody), &body); err != nil {
		kong.Log.Err("Failed to unmarshal body: ", err.Error())
		return
	}

	if body == nil {
		body = make(map[string]interface{})
	}
	header, ok := body["Header"].(map[string]interface{})
	if !ok {
		header = make(map[string]interface{})
		body["Header"] = header
	}
	header["payor_code"] = payorCode

	newRawBody, err := json.Marshal(body)
	if err != nil {
		kong.Log.Err("Failed to marshal body: ", err.Error())
		return
	}

	if err := kong.ServiceRequest.SetRawBody(string(newRawBody)); err != nil {
		kong.Log.Err("Failed to set raw body: ", err.Error())
	}
}
