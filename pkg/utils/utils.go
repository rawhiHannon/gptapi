package utils

import (
	_ "encoding/json"
	"log"
	"os"
	"path/filepath"

	// "errors"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	// "log"
	"net/http"
	"strconv"

	// "strings"
	"time"
)

type HttpHandler = func(http.ResponseWriter, *http.Request)

type ChannelPair map[string]chan struct{}

type Middleware interface {
	PreProcessor(http.ResponseWriter, *http.Request) error
	PostProcessor(map[string]chan struct{}, string) error
}

type Action struct {
	Route        string
	Handler      RequestHandler
	Midddlewares []Middleware
	priority     int
}

type RequestHandler = func(map[string]string, map[string][]string, map[string]interface{}) (string, error)

func ExtractBody(req *http.Request) map[string]interface{} {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	var anyJson map[string]interface{}
	err = json.Unmarshal(body, &anyJson)
	if err != nil {
		//TODO:
		// log.Println(err, anyJson)
	}
	return anyJson
}

func SendHttpResponse(w http.ResponseWriter, success bool, response string, responseKey string) {
	w.Header().Set("Content-Type", "application/json")
	if responseKey == "" {
		responseKey = "data"
	}
	if success == false {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": "fail", "%s": "%s"}`, responseKey, response)))
	} else {
		w.WriteHeader(http.StatusOK)
		if response != "" {
			w.Write([]byte(fmt.Sprintf(`{"status": "ok", "%s": %s}`, responseKey, response)))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"status": "ok"}`)))
		}
	}
}

func GetHttpWrapper(handler RequestHandler, middlewares []Middleware) func(http.ResponseWriter, *http.Request) {
	return (func(w http.ResponseWriter, r *http.Request) {
		body := ExtractBody(r)
		params := mux.Vars(r)
		queryString := r.URL.Query()
		//TODO: handle priority
		//TODO: handle return immediate results via middlewares instead of err
		if middlewares != nil && len(middlewares) > 0 {
			for _, middleware := range middlewares {
				err := middleware.PreProcessor(w, r)
				if err != nil {
					SendHttpResponse(w, false, err.Error(), "")
					return
				}
			}
		}
		data, err := handler(params, queryString, body)
		if err != nil {
			SendHttpResponse(w, false, err.Error(), "")
			return
		}
		SendHttpResponse(w, true, data, "")
	})
}

func MultiplyDuration(factor int, d time.Duration) time.Duration {
	return time.Duration(factor) * d // method 1 -- multiply in 'Duration'
	// return time.Duration(factor * int64(d)) // method 2 -- multiply in 'int64'
}

func IsInt(data string) bool {
	_, ok := strconv.Atoi(data)
	return (ok == nil)
}

func ToJSON(data map[string]string) string {
	out, _ := json.Marshal(data)
	return string(out)
}

func StringArrayToJSON(data []string) string {
	out, _ := json.Marshal(data)
	return string(out)
}

func MapArrayToJSON(data []map[string]interface{}) string {
	out, _ := json.Marshal(data)
	return string(out)
}

func ArrayToJSON(data []interface{}) string {
	out, _ := json.Marshal(data)
	return string(out)
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func ToString(data interface{}) string {
	if data == nil {
		return ""
	}
	switch data.(type) {
	case string:
		return data.(string)
	case float64:
		fmt.Println(data)
		return fmt.Sprint(data)
	default:
		fmt.Println(data)
		return ""
	}
}

func LoadEnv(path string) {
	absPath, _ := filepath.Abs(path + ".env")
	if path == "" {
		absPath = ".env"
	}
	err := godotenv.Load(absPath)
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}
}
