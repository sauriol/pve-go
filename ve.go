package proxmox

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

type Proxmox struct {
	Hostname            string
	Username            string
	password            string
	SSL                 bool
	APIPath             string
	CSRFPreventionToken string
	Ticket              string
	Client              *http.Client
}

func (proxmox Proxmox) action(method string,
	endpoint string,
	vals url.Values) (interface{}, error) {
	var req *http.Request
	var err error

	target := proxmox.APIPath + endpoint

	switch method {
	case "POST", "PUT":
		if vals == nil {
			return nil, errors.New("Data must not be nil")
		}

		req, err = http.NewRequest(method, target,
			bytes.NewBufferString(vals.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(vals.Encode())))
		if proxmox.CSRFPreventionToken != "" {
			req.Header.Add("CSRFPreventionToken", proxmox.CSRFPreventionToken)
		}
	case "GET":
		req, err = http.NewRequest(method, target, nil)
		if err != nil {
			return nil, err
		}
	case "DELETE":
		req, err = http.NewRequest(method, target, nil)
		if err != nil {
			return nil, err
		}

		if proxmox.CSRFPreventionToken != "" {
			req.Header.Add("CSRFPreventionToken", proxmox.CSRFPreventionToken)
		}
	default:
		return nil, errors.New("Invalid method: \"" + method + "\"")
	}

	resp, err := proxmox.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("HTTP Error: " + resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var body map[string]interface{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return nil, err
	}
	if body == nil || body["data"] == nil {
		return nil,
			errors.New("Invalid response to " + method + " to " + target)
	}

	m, mapOk := body["data"].(map[string]interface{})
	a, arrOk := body["data"].([]interface{})

	if mapOk {
		return m, nil
	} else if arrOk {
		return a, nil
	} else {
		return nil, errors.New("Invalid response to " + method + " to " + target)
	}
}

func (proxmox Proxmox) PostForm(endpoint string,
	form url.Values) (interface{}, error) {

	if endpoint != "" && form != nil {
		data, err := proxmox.action("POST", endpoint, form)
		if err != nil {
			return nil, err
		}

		return data, nil
	} else {
		return nil, errors.New("Invalid parameters passed to PostForm")
	}
}

func (proxmox Proxmox) PutForm(endpoint string,
	form url.Values) (interface{}, error) {

	if endpoint != "" && form != nil {
		data, err := proxmox.action("PUT", endpoint, form)
		if err != nil {
			return nil, err
		}

		return data, nil
	} else {
		return nil, errors.New("Invalid parameters passed to PutForm")
	}
}

func (proxmox Proxmox) Get(endpoint string) (interface{}, error) {
	if endpoint != "" {
		data, err := proxmox.action("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		return data, nil
	} else {
		return nil, errors.New("Invalid parameters passed to Get")
	}
}

func (proxmox Proxmox) Delete(endpoint string) (interface{}, error) {
	if endpoint != "" {
		data, err := proxmox.action("DELETE", endpoint, nil)
		if err != nil {
			return nil, err
		}

		return data, nil
	} else {
		return nil, errors.New("Invalid parameters passed to Delete")
	}
}

func InitProxmox(Hostname string,
	Username string,
	Password string) (*Proxmox, error) {

	if !strings.HasPrefix(Hostname, "http") &&
		!strings.HasPrefix(Hostname, "https") {
		Hostname = "https://" + Hostname
	}

	if !strings.Contains(Username, "@") {
		Username += "@pam"
	}

	proxmox := new(Proxmox)
	proxmox.Hostname = Hostname
	proxmox.Username = Username
	proxmox.password = Password
	proxmox.SSL = true

	if len(strings.Split(proxmox.Hostname, ":")) == 2 {
		proxmox.APIPath = proxmox.Hostname + ":8006/api2/json"
	} else {
		proxmox.APIPath = proxmox.Hostname + "/api2/json"
	}

	tr := &http.Transport{
		DisableKeepAlives:   false,
		IdleConnTimeout:     0,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: proxmox.SSL},
	}

	proxmox.Client = &http.Client{Transport: tr}

	form := url.Values{
		"username": {proxmox.Username},
		"password": {proxmox.password},
	}

	// Authenticate to the server
	data, err := proxmox.PostForm("/access/ticket", form)
	dataMap := data.(map[string]interface{})
	if err != nil {
		return nil, err
	}

	// Set the ticket and token
	proxmox.Ticket = dataMap["ticket"].(string)
	proxmox.CSRFPreventionToken = dataMap["CSRFPreventionToken"].(string)

	proxmox.Client.Jar, err = cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:  "PVEAuthCookie",
		Value: proxmox.Ticket,
		Path:  "/",
	}
	cookies = append(cookies, cookie)
	url, err := url.Parse(proxmox.Hostname + "/")
	if err != nil {
		return nil, err
	}
	proxmox.Client.Jar.SetCookies(url, cookies)

	return proxmox, nil
}
