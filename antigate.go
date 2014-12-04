package antigate

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type LoadStat struct {
	Waiting                  int     `xml:"waiting"`
	WaitingRU                int     `xml:"waitingRU"`
	Load                     float32 `xml:"load"`
	Minbid                   float64 `xml:"minbid"`
	MinbidRU                 float64 `xml:"minbidRU"`
	AverageRecognitionTime   float64 `xml:"averageRecognitionTime"`
	AverageRecognitionTimeRU float64 `xml:"averageRecognitionTimeRU"`
}

type Account struct {
	ApiKey string
}

const (
	BaseURL = "http://anti-captcha.com"
)

func New(api_key string) Account {
	return Account{ApiKey: api_key}
}

func (a Account) UploadImage(file_bytes []byte) (int, error) {
	body := base64.StdEncoding.EncodeToString(file_bytes)
	url_api := BaseURL + "/in.php"
	params := url.Values{
		"key":    {a.ApiKey},
		"method": {"base64"},
		"body":   {body},
	}
	resp, err_post := http.PostForm(url_api, params)
	if err_post != nil {
		return 0, err_post
	}
	defer resp.Body.Close()
	data, err_read := ioutil.ReadAll(resp.Body)
	if err_read != nil {
		return 0, err_read
	}
	res := string(data)

	if strings.HasPrefix(res, "OK|") {
		captcha_id, err_atoi := strconv.Atoi(res[3:])
		if err_atoi != nil {
			return 0, err_atoi
		}
		return captcha_id, nil
	} else {
		return 0, errors.New(res)
	}
}

func (a Account) CheckStatus(captcha_id int) (bool, string, error) {
	url_api := fmt.Sprintf("%s/res.php?key=%s&action=get&id=%d", BaseURL, a.ApiKey, captcha_id)
	resp, err_get := http.Get(url_api)
	if err_get != nil {
		return false, "", err_get
	}
	defer resp.Body.Close()
	data, err_read := ioutil.ReadAll(resp.Body)
	if err_read != nil {
		return false, "", err_read
	}
	res := string(data)
	fmt.Println(res)
	if strings.HasPrefix(res, "OK|") {
		return true, res[3:], nil
	} else if res == "CAPCHA_NOT_READY" {
		return false, "", nil
	} else {
		return false, "", errors.New(res)
	}
}

func (a Account) GetBalance() (float64, error) {
	url_balance := fmt.Sprintf("%s/res.php?key=%s&action=getbalance", BaseURL, a.ApiKey)
	resp, err_get := http.Get(url_balance)
	if err_get != nil {
		return 0.0, err_get
	}
	defer resp.Body.Close()
	data, err_read := ioutil.ReadAll(resp.Body)
	if err_read != nil {
		return 0.0, err_read
	}
	balance, err_parse := strconv.ParseFloat(string(data), 64)
	if err_parse != nil {
		return 0.0, err_parse
	}
	return balance, nil
}

func GetSystemStat() (*LoadStat, error) {
	var load_stat *LoadStat
	url_stat := BaseURL + "/load.php"
	resp, err_get := http.Get(url_stat)
	if err_get != nil {
		return load_stat, err_get
	}
	defer resp.Body.Close()
	data, err_read := ioutil.ReadAll(resp.Body)
	if err_read != nil {
		return load_stat, err_read
	}
	err_unmarshal := xml.Unmarshal(data, &load_stat)

	if err_unmarshal != nil {
		return load_stat, err_unmarshal
	}
	return load_stat, nil
}
