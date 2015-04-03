package utils

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Arg(req *http.Request, key string) (string, bool) {
	param := req.URL.Query()[key]
	if param != nil && len(param) > 0 {
		return param[0], true
	}
	return "", false
}

func ArgDef(req *http.Request, key string, def string) string {
	param, found := Arg(req, key)
	if !found {
		return def
	}
	return param
}

func ArgInt(req *http.Request, key string) (int, bool) {
	param, found := Arg(req, key)
	bigInt, err := strconv.ParseInt(param, 10, 0)
	intVal := int(bigInt)
	if !found || err != nil {
		return 0, false
	}
	return intVal, true
}

func ArgIntDef(req *http.Request, key string, def int) int {
	param, found := ArgInt(req, key)
	if !found {
		return def
	}
	return param
}

func ArgBool(req *http.Request, key string) bool {
	_, found := Arg(req, key)
	return found
}

func FormDef(r *http.Request, key string, def string) string {
	val := r.FormValue(key)
	if val == "" {
		return def
	}
	return val
}

func GetHeader(req *http.Request, name string) (string, bool) {
	upgrade := req.Header[name]
	if upgrade == nil || len(upgrade) <= 0 {
		return "", false
	}
	return upgrade[0], true
}

func IsWebsocketReq(req *http.Request) bool {
	header, _ := GetHeader(req, "Upgrade")
	return header == "websocket"
}

func IsHtmlReq(req *http.Request) bool {
	header, _ := GetHeader(req, "Accept")
	return strings.Contains(header, "/html")
}

func IsJsonReq(req *http.Request) bool {
	header, _ := GetHeader(req, "Accept")
	return strings.Contains(header, "/json")
}

func HasFormData(r *http.Request) bool {
	header, _ := GetHeader(r, "Content-Type")
	return strings.Contains(header, "multipart/form-data")
}

func Error(w http.ResponseWriter, r *http.Request, message string, err error) {
	log.Println("ERROR: ", message)
	log.Println(err)
	http.Error(w, message, http.StatusInternalServerError)
}

func Flush(w io.Writer) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
