package weedCL

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	addr string
)

func assign(rw http.ResponseWriter, req *http.Request) {
	out := fmt.Sprintf(`{"count":1,"fid":"3,01637037d6","url":"%s","publicUrl":"%s"}`, addr, addr)
	rw.Write([]byte(out))
}

func put(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{"size": 43234}`))
}

func lookup(rw http.ResponseWriter, req *http.Request) {
	out := fmt.Sprintf(`{"locations":[{"publicUrl":"%s","url":"%s"}]}`, addr, addr)
	rw.Write([]byte(out))
}

func setupMockWeedfs() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/dir/assign", assign)
	mux.HandleFunc("/dir/lookup", lookup)
	mux.HandleFunc("/", put)
	ts := httptest.NewServer(mux)
	addr = ts.URL
	return ts
}

func testClient(addr string) *Client {
	cfg := NewHTTPCfg(addr)
	return NewClient(cfg)
}

func TestMain(m *testing.M) {
	ts := setupMockWeedfs()
	fmt.Println("test sever started at", ts.URL)
	defer ts.Close()

	os.Exit(m.Run())
}

func TestAssign(t *testing.T) {
	cl := testClient(addr)
	resp, err := cl.assign()
	assert.NoError(t, err, "assign err")

	assert.Equal(t, 1, resp.Count, "count doesn't match")
	assert.Equal(t, "3,01637037d6", resp.Fid, "fid doesn't match")
	assert.Equal(t, addr, resp.PublicURL, "publicURL doesn't match")
	assert.Equal(t, addr, resp.URL, "url doesn't match")
	assert.Equal(t, fmt.Sprintf("%s/%s", resp.URL, resp.Fid), resp.PutURL(), "putURL doesn't match")
}

func TestLookup(t *testing.T) {
	cl := testClient(addr)
	resp, err := cl.lookupFid("3,01637037d6")
	assert.NoError(t, err, "assign err")

	assert.Len(t, resp.Locations, 1, "len of locations should be 1")
	assert.Equal(t, addr, resp.Locations[0].PublicURL, "publicURL doesn't match")
	assert.Equal(t, addr, resp.Locations[0].URL, "url doesn't match")
}

func TestUpload(t *testing.T) {
	cl := testClient(addr)
	buff := &bytes.Buffer{}
	wr := bufio.NewWriter(buff)
	wr.WriteString("test")

	resp, err := cl.Upload("test", "application/octet-stream", buff)
	assert.NoError(t, err)
	assert.Equal(t, "3,01637037d6", resp)
}

func TestDownload(t *testing.T) {
	cl := testClient(addr)
	resp, err := cl.Download("3,01637037d6")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
