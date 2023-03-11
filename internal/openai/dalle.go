package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	apiKey     string
	baseURL    string
	userAgent  string
	userId     string
	timeOut    time.Duration // 超时时间, 0表示不超时
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

// NewDallE 新建一个智能绘图
func NewDallE(ApiKey, UserId string, timeOut time.Duration) *DallE {
	httpClient := &http.Client{
		Timeout: defaultTimeout,
	}
	return &DallE{
		apiKey:     ApiKey,
		baseURL:    "https://api.openai.com/v1/images",
		userAgent:  "go-dalle",
		userId:     UserId,
		timeOut:    timeOut,
		httpClient: httpClient,
	}
}

func (d *DallE) GenPhoto(prompt string, n int, size string) ([]string, error) {
	if len(size) == 0 {
		size = "512x512"
	}

	requestBody := DallEReq{
		Prompt:         prompt,
		N:              n,
		Size:           size,
		ResponseFormat: "url",
		// Determinism:    0.5,
	}

	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewReader(postData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", d.apiKey))
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, e
	}
	// mylog.Debug(string(body))

	var dallEResp DallEResp
	err = json.Unmarshal(body, &dallEResp)
	if err != nil {
		return nil, err
	}

	if len(dallEResp.Error.Message) > 0 {
		return nil, fmt.Errorf("%v", dallEResp.Error.Message)
	}

	var out []string
	for _, v := range dallEResp.Data {
		out = append(out, v.Url)
	}
	return out, nil
}

func (d *DallE) GenPhotoBase64(prompt string, n int, size string) ([]string, error) {
	if len(size) == 0 {
		size = "512x512"
	}

	requestBody := DallEReq{
		Prompt:         prompt,
		N:              n,
		Size:           size,
		ResponseFormat: "b64_json",
	}

	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewReader(postData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", d.apiKey))
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, e
	}
	// mylog.Debug(string(body))

	var dallEResp DallEResp
	err = json.Unmarshal(body, &dallEResp)
	if err != nil {
		return nil, err
	}

	if len(dallEResp.Error.Message) > 0 {
		return nil, fmt.Errorf("%v", dallEResp.Error.Message)
	}

	var out []string
	for _, v := range dallEResp.Data {
		out = append(out, v.B64Json)
	}
	return out, nil
}

func pointerizeString(s string) *string {
	return &s
}

func (c *DallE) Edit(prompt string, image *os.File, mask *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	url := c.baseURL + "/edits"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if image == nil {
		return nil, errors.New("image is nil")
	}
	if mask == nil {
		return nil, errors.New("mask is nil")
	}
	log.Println(image.Name())

	if imageWriter, err := w.CreateFormFile("image", image.Name()); err != nil {
		return nil, err
	} else if _, err := io.Copy(imageWriter, image); err != nil {
		return nil, err
	}
	if maskWriter, err := w.CreateFormFile("mask", mask.Name()); err != nil {
		return nil, err
	} else if _, err := io.Copy(maskWriter, mask); err != nil {
		return nil, err
	}
	var sizeStr *string
	if size != nil {
		sizeStr = pointerizeString(fmt.Sprintf("%dx%d", size, size))
	}
	err := w.WriteField("prompt", prompt)
	if err != nil {
		return nil, err
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
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println(resp)
	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("bad request")
		case 401:
			return nil, errors.New("unauthorized")
		case 403:
			return nil, errors.New("forbidden")
		case 404:
			return nil, errors.New("not found")
		case 429:
			return nil, errors.New("too many requests")
		case 500:
			return nil, errors.New("internal server error")
		case 502:
			return nil, errors.New("bad gateway")
		case 503:
			return nil, errors.New("service unavailable")
		case 504:
			return nil, errors.New("gateway timeout")
		default:
			return nil, errors.New("unknown error")
		}
	}
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// Image is the image to edit.
//
// Size is the size of the image to generate (Small, Medium, Large).
//
// N is the number of images to generate.
//
// https://beta.openai.com/docs/guides/images/variations
func (c *DallE) Variation(image *os.File, size *int, n *int, user *string, responseType *string) ([]Datum, error) {
	url := c.baseURL + "/variations"

	// this is posting using multipart/form-data

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

	req, err := http.NewRequest("POST", url, &b)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("bad request")
		case 401:
			return nil, errors.New("unauthorized")
		case 403:
			return nil, errors.New("forbidden")
		case 404:
			return nil, errors.New("not found")
		case 429:
			return nil, errors.New("too many requests")
		case 500:
			return nil, errors.New("internal server error")
		case 502:
			return nil, errors.New("bad gateway")
		case 503:
			return nil, errors.New("service unavailable")
		case 504:
			return nil, errors.New("gateway timeout")
		default:
			return nil, errors.New("unknown error")
		}
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}
