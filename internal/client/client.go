package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"relay/internal/models"
	"relay/internal/server"
)

type Client struct {
	url string
}

func NewClient(port int) Client {
	return Client{fmt.Sprintf("http://localhost:%d", port)}
}

func (c Client) Ping() error {
	resp, err := http.Get(fmt.Sprintf("%s/", c.url))
	if err != nil {
		return fmt.Errorf("Ping: http.Get %+w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Ping: io.ReadAll %+w", err)
	}
	fmt.Printf("Ping: msg recieved: %s\n", out)
	return nil
}

func (c Client) GetRelay(id int) (*models.Relay, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/relays/%d", c.url, id))
	if err != nil {
		return nil, fmt.Errorf("GetRelay: http.Get %+w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetRelay: io.ReadAll %+w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GetRelay: response status: %d, body %s", resp.StatusCode, body)
	}

	var relay models.Relay
	err = json.Unmarshal(body, &relay)
	if err != nil {
		return nil, fmt.Errorf("GetRelay: json.Unmarshal: %+w", err)
	}

	return &relay, nil
}

func (c Client) CreateRelay(deviceId string, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime *time.Time) (int, error) {
	data := server.CreateRelayRequest{
		DeviceId:          deviceId,
		CloudFunction:     cloudFunction,
		Argument:          &argument,
		DesiredReturnCode: desiredReturnCode,
		ScheduledTime:     scheduledTime,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: json.Marshal: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/relays", c.url), bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: io.ReadAll: %w", err)
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("CreateRelay: request error, status code=%d, body=%s", resp.StatusCode, body)
	}

	id, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: strconv.ParseInt: %w on %s", err, string(body))
	}
	return int(id), nil
}

func (c Client) CancelRelay(id int) error {
	req, err := http.NewRequest("Delete", fmt.Sprintf("%s/api/relays/%d", c.url, id), nil)
	if err != nil {
		return fmt.Errorf("CancelRelay: http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("CancelRelay: client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("CancelRelay: io.ReadAll: %w", err)
	}

	// Check the response status code
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("CancelRelay: request error, status code=%d, body=%s", resp.StatusCode, body)
	}
	return nil
}
