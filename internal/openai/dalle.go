package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	Small  int = 256
	Medium int = 512
	Large  int = 1024
)

const (
	defaultBaseURL   = "https://api.openai.com/v1/images"
	defaultUserAgent = "go-dalle"
	defaultTimeout   = 30 * time.Second
)

type Response struct {
	Created int64   `json:"created"`
	Data    []Datum `json:"data"`
}

type Datum struct {
	URL string `json:"url"`
}

type DallE struct {
	id         uint64
	apiKey     string
	baseURL    string
	userAgent  string
	userId     string
	timeOut    time.Duration
	httpClient *http.Client
}

type DallEReq struct {
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format"`
	// Determinism    float32 `json:"determinism"`
}

type DallEResp struct {
	Created int64       `json:"created"`
	Data    []DallEData `json:"data"`
	Error   DallError   `json:"error"`
}

type DallEData struct {
	Url     string `json:"url"`
	B64Json string `json:"b64_json"`
}

type DallError struct {
	Message string `json:"message"`
}

func NewDallE(ApiKey string, id uint64, limit int, rate time.Duration) *DallE {
	httpClient := &http.Client{
		Timeout: defaultTimeout,
	}
	d := &DallE{
		id:         id,
		apiKey:     ApiKey,
		baseURL:    "https://api.openai.com/v1/images",
		userAgent:  "go-dalle",
		httpClient: httpClient,
	}
	return d
}

func pointerizeString(s string) *string {
	return &s
}

func extractError(resp *http.Response) error {
	switch resp.StatusCode {
	case 400:
		return errors.New("bad request")
	case 401:
		return errors.New("unauthorized")
	case 403:
		return errors.New("forbidden")
	case 404:
		return errors.New("not found")
	case 429:
		return errors.New("too many requests")
	case 500:
		return errors.New("internal server error")
	case 502:
		return errors.New("bad gateway")
	case 503:
		return errors.New("service unavailable")
	case 504:
		return errors.New("gateway timeout")
	default:
		return errors.New("unknown error")
	}
}

func (d *DallE) sendRequest(endpoint string, contentType string, requestBody, responsePtr interface{}) error {
	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", fmt.Sprintf(`%s/%s`, defaultBaseURL, endpoint), bytes.NewReader(postData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", contentType)
	req.Header.Set("User-Agent", d.userAgent)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", d.apiKey))
	resp, e := client.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return extractError(resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return e
	}
	err = json.Unmarshal(body, responsePtr)
	if err != nil {
		return err
	}
	return nil
}

func (d *DallE) createPhoto(prompt string, n int, size string, responseFormat string, contentType string) ([]string, error) {
	if responseFormat != "url" && responseFormat != "b64_json" {
		responseFormat = "url"
	}
	if len(size) == 0 {
		size = "512x512"
	}
	requestBody := DallEReq{
		Prompt:         prompt,
		N:              n,
		Size:           size,
		ResponseFormat: responseFormat,
		// Determinism:    0.5,
	}
	var dallEResp DallEResp
	err := d.sendRequest("generations", contentType, requestBody, &dallEResp)
	if err != nil {
		return []string{}, err
	}
	var out []string
	for _, v := range dallEResp.Data {
		out = append(out, v.Url)
	}
	return out, nil
}

func (d *DallE) GenPhoto(prompt string, n int, size string) ([]string, error) {
	return d.createPhoto(prompt, n, size, "url", "application/json;charset=utf-8")
}

func (d *DallE) GenPhotoBase64(prompt string, n int, size string) ([]string, error) {
	return d.createPhoto(prompt, n, size, "b64_json", "application/json;charset=utf-8")
}

func (d *DallE) Variation(image *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if image == nil {
		return nil, errors.New("image is nil")
	}
	imageWriter, err := w.CreateFormFile("image", image.Name())
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(imageWriter, image); err != nil {
		return nil, err
	}
	var sizeStr *string
	if size != nil {
		sizeStr = pointerizeString(fmt.Sprintf("%dx%d", size, size))
	}
	if n != nil {
		err = w.WriteField("n", fmt.Sprintf("%d", n))
		if err != nil {
			return nil, err
		}
	}
	if sizeStr != nil {
		err = w.WriteField("size", *sizeStr)
		if err != nil {
			return nil, err
		}
	}
	if user != nil {
		err = w.WriteField("user", *user)
		if err != nil {
			return nil, err
		}
	}
	if responseType != nil {
		err = w.WriteField("response_format", *responseType)
		if err != nil {
			return nil, err
		}
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	var response Response
	d.sendRequest("variations", w.FormDataContentType(), &b, &response)
	return response.Data, nil
}
