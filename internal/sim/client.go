package sim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"waze/internal/types"
)

type Client struct {
	BaseURL string
	Http    *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		BaseURL: url,
		Http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// send all traffic report from al cars to server
func (c *Client) SendTrafficBatch(reports []types.TrafficReport) error {
	jsonData, _ := json.Marshal(reports)
	resp, err := c.Http.Post(c.BaseURL+"/api/traffic", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// close the connetion after the function ends
	defer resp.Body.Close()

	// check connection status
	if resp.StatusCode != 200 {
		return fmt.Errorf("Server returned status:  %d", resp.StatusCode)
	}
	return nil
}

// Request and return route from server
func (c *Client) RequestRoute(startNode, endNode int) ([]int, error) {

	url := fmt.Sprintf("%s/api/navigate?from=%d&to=%d", c.BaseURL, startNode, endNode)
	// fmt.Println(url)
	// time.Sleep(time.Second * 5)

	// send route request
	resp, err := c.Http.Get(url)
	if err != nil {
		return nil, err
	}
	// close the connection after function ends
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("navigation failed, status: %d", resp.StatusCode)
	}

	var result types.NavigationResponse

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.RouteNodes, nil
}
