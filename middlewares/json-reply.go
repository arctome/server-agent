package Middlewares

import "encoding/json"

func Json2String(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
