package antigate

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client - AntiGate Client.
type Client struct {
	URL string
	Key string

	CheckInterval time.Duration // solve check interval
	RetryInterval time.Duration // retry interval when too much traffic
	MaxRetries    int
}

const (
	// BaseURL - AntiGate base url.
	BaseURL = "http://anti-captcha.com"
)

var (
	// ErrNoSlots - Error returned when API has no available workers.
	ErrNoSlots = errors.New("No slots available")

	// ErrCaptchaNotReady - Error returned when captcha is work in progress.
	ErrCaptchaNotReady = errors.New("Captcha is not yet solved")

	// ErrorNames - Errors by name returned from API.
	ErrorNames = map[string]error{
		"ERROR_NO_SLOT_AVAILABLE": ErrNoSlots,
		"CAPCHA_NOT_READY":        ErrCaptchaNotReady,
	}
)

// New - Creates new AntiGate client.
// Default values:
// 	CheckInterval - 2.5 seconds
// 	RetryInterval - 2.5 seconds
// 	MaxRetries    - 35
func New(key string) *Client {
	return &Client{
		Key:           key,
		MaxRetries:    35,
		CheckInterval: 2500 * time.Millisecond,
		RetryInterval: 2500 * time.Millisecond,
	}
}

// GetError - Gets Antigate error by name.
func GetError(text string) error {
	err, ok := ErrorNames[text]
	if ok {
		return err
	}
	return errors.New(text)
}

// GetErrorBytes - Gets Antigate error by name.
func GetErrorBytes(text []byte) error {
	return GetError(string(text))
}

var captchaOKPrefix = []byte("OK|")

// Solve - Solves captcha.
func (client *Client) Solve(image []byte) (result string, err error) {
	// Upload captcha
	captcha, err := client.UploadImage(image)
	if err != nil {
		return
	}

	// Start ticking every client.CheckInterval
	for _ = range time.Tick(client.CheckInterval) {
		var ok bool
		ok, result, err = client.GetStatus(captcha)
		if ok || err != nil {
			return
		}
	}

	return
}

// UploadImage - Uploads image to AntiGate API.
func (client *Client) UploadImage(image []byte) (captcha int, err error) {
	return client.uploadImage(0, image)
}

// GetStatus -
func (client *Client) GetStatus(captcha int) (ok bool, result string, err error) {
	return client.getStatus(0, captcha)
}

// GetBalance - Gets account balance.
func (client *Client) GetBalance() (balance float64, err error) {
	// Get balance from API
	resp, err := http.Get(client.GetURL("/res.php?key=%s&action=getbalance", client.Key))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Parse float and return
	balance, err = strconv.ParseFloat(string(data), 64)
	return
}

func (client *Client) uploadImage(retry int, image []byte) (captcha int, err error) {
	// Encode image body with base64
	body := base64.StdEncoding.EncodeToString(image)
	params := url.Values{
		"key":    {client.Key},
		"method": {"base64"},
		"body":   {body},
	}

	// Upload image
	resp, err := http.PostForm(client.GetURL("/in.php"), params)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// If response starts with OK|
	// Parse the rest as captcha ID
	if bytes.HasPrefix(data, captchaOKPrefix) {
		return strconv.Atoi(string(data[3:]))
	}

	// Create error from response body otherwise
	err = GetErrorBytes(data)

	if err == ErrNoSlots && retry < client.MaxRetries {
		<-time.After(client.RetryInterval)
		return client.uploadImage(retry+1, image)
	}

	return
}

func (client *Client) getStatus(retry int, captcha int) (ok bool, result string, err error) {
	// Get status from API
	resp, err := http.Get(client.GetURL("/res.php?key=%s&action=get&id=%d", client.Key, captcha))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// If response has prefix OK captcha is solved
	// Get rest of the message as captcha text
	if ok = bytes.HasPrefix(data, captchaOKPrefix); ok {
		result = string(data[3:])
		return
	}

	// Make error from response data
	err = GetErrorBytes(data)

	// If err is captcha not ready return empty error
	if err == ErrCaptchaNotReady {
		err = nil
	}

	if err == ErrNoSlots && retry < client.MaxRetries {
		<-time.After(client.RetryInterval)
		return client.getStatus(retry+1, captcha)
	}

	return
}

// GetURL - Gets full api url.
func (client *Client) GetURL(path string, v ...interface{}) string {
	if client.URL == "" {
		client.URL = BaseURL
	}
	if len(v) > 0 {
		path = fmt.Sprintf(path, v...)
	}
	return fmt.Sprintf("%s%s", client.URL, path)
}

// LoadStat - AntiGate load statistics.
type LoadStat struct {
	Waiting                  int     `xml:"waiting"`
	WaitingRU                int     `xml:"waitingRU"`
	Load                     float32 `xml:"load"`
	Minbid                   float64 `xml:"minbid"`
	MinbidRU                 float64 `xml:"minbidRU"`
	AverageRecognitionTime   float64 `xml:"averageRecognitionTime"`
	AverageRecognitionTimeRU float64 `xml:"averageRecognitionTimeRU"`
}

// GetSystemStat - Gets AntiGate load statistics.
func GetSystemStat() (stats *LoadStat, err error) {
	// Get stats from API
	resp, err := http.Get(BaseURL + "/load.php")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Unmarshal XML and return
	err = xml.Unmarshal(data, &stats)
	return
}
