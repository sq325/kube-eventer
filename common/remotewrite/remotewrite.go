package remotewrite

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/golang/snappy"

	"github.com/sq325/kube-eventer/common/remotewrite/prompb"

	"google.golang.org/protobuf/proto"
)

type RemoteWriteClient interface {
	Name() string
	Write([]*prompb.TimeSeries) error
}

type RemoteWriteOption struct {
}

type Client struct {
	url    string
	client *http.Client
}

func NewClient(url string) *Client {
	dialTimeout := time.Duration(5 * time.Second)
	timeout := time.Duration(10 * time.Second)
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: dialTimeout,
		}).DialContext,
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   100,
	}
	httpclient := &http.Client{
		Transport: tr,
	}

	return &Client{
		url:    url,
		client: httpclient,
	}
}

func (c *Client) Name() string {
	return "RemoteWrite Client"
}

func (c *Client) Write(series []*prompb.TimeSeries) error {
	if len(series) == 0 {
		return nil
	}

	req := &prompb.WriteRequest{
		Timeseries: series,
	}

	bys, err := proto.Marshal(req)
	if err != nil {
		log.Printf("failed to marshal WriteRequest: %v", err)
		return err
	}
	if err := c.write(snappy.Encode(nil, bys)); err != nil {
		return err
	}
	return nil
}

func (c *Client) write(bys []byte) error {
	req, err := http.NewRequest("POST", c.url, bytes.NewReader(bys))
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return err
	}
	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("User-Agent", "kube-eventer")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Println("push data with remote write request got error:", err, "response body:", string(bys))
		return err
	}
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("push data with remote write request got status code: %v, response body: %s", resp.StatusCode, string(bys))
		return err
	}

	return nil
}
