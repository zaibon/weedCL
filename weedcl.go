package weedCL

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"io"
)

type Client struct {
	client *http.Client
	Cfg    *HTTPConfig
}

func NewClient(cfg *HTTPConfig) *Client {
	return &Client{
		client: http.DefaultClient,
		Cfg:    cfg,
	}
}

//Upload uploads a file  on weedfs and returns the fid of weedfs and an error if any
//filename is the name of the file
//contentType is the mime type of the file
//body is a reader on the file to upload
func (c *Client) Upload(filename, contentType string, body io.Reader) (string, error) {
	assignRsp, err := c.assign()
	if err != nil {
		return "", err
	}

	mp, contentType, err := createMultiPart(filename, body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", assignRsp.PublicURL, mp)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bResp, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("error during upload (%s) :%s", resp.StatusCode, string(bResp))
	}
	return assignRsp.Fid, nil
}

//Download make a request on weeds and return a io.ReadCloser
// that contain the file identified by fid or an error if any
func (c *Client) Download(fid string) (io.ReadCloser, error) {
	lookupResp, err := c.lookupFid(fid)
	if err != nil {
		return nil, err
	}
	volURL := ""
	for _, loc := range lookupResp.Locations {
		if loc.PublicURL != "" {
			volURL = loc.PublicURL
			break
		}
	}
	if volURL == "" {
		return nil, fmt.Errorf("no volume url found for that fid")
	}
	resp, err := c.client.Get(fmt.Sprintf("%s/%s", volURL, fid))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// AssignResponse is the response for an assign (new file upload) such as
// {"count":1,"fid":"3,01637037d6","url":"127.0.0.1:8080","publicUrl":"localhost:8080"}
type assignRsp struct {
	Count     int    `json:"count"`
	Fid       string `json:"fid"`
	URL       string `json:"url"`
	PublicURL string `json:"publicUrl"`
}

// assign get a new fileID from master
func (c *Client) assign() (*assignRsp, error) {
	u := fmt.Sprintf("%s/dir/assign", c.Cfg.BaseURL.String())
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	assignRsp := &assignRsp{}
	err = json.NewDecoder(resp.Body).Decode(assignRsp)
	return assignRsp, err
}

func createMultiPart(filename string, body io.Reader) (io.Reader, string, error) {
	buff := &bytes.Buffer{}
	mpW := multipart.NewWriter(buff)
	defer mpW.Close()
	filePart, err := mpW.CreateFormFile("file", filename)

	_, err = io.Copy(filePart, body)
	if err != nil {
		return nil, "", err
	}
	return buff, mpW.FormDataContentType(), nil
}

func getVolID(fid string) string {
	var volID string
	if i := strings.Index(fid, ","); i > 0 {
		volID = fid[:i]
	} else {
		volID = fid
	}
	return volID
}

type lookupResp struct {
	Locations []struct {
		PublicURL string `json:"publicUrl"`
		URL       string `json:"url"`
	} `json:"locations"`
}

//lookupFid does a lookup request on weedfs for fid and return a lookupResp or an error if any
func (c *Client) lookupFid(fid string) (*lookupResp, error) {
	volID := getVolID(fid)
	u := fmt.Sprintf("%s/dir/lookup?volumeId=%s", c.Cfg.BaseURL.String(), volID)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	lookupResp := &lookupResp{}
	if err := json.NewDecoder(resp.Body).Decode(lookupResp); err != nil {
		return nil, err
	}
	return lookupResp, nil
}
