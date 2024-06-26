package particle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ParticleAPI interface {
	Ping(deviceId string) (bool, error)
	CloudFunction(deviceId string, cloudFunction string, argument string, returnValue *int) (bool, error)
}

type Particle struct {
	token string
}

func NewParticle(token string) (*Particle, error) {
	valid, err := testToken(token)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("NewParticle: invalid particle token")
	}
	return &Particle{token: token}, nil
}

// https://docs.particle.io/reference/cloud-apis/api/#errors
// 408 is not actually used?

func (p Particle) Ping(deviceId string) (bool, error) {
	queryParams := url.Values{}
	queryParams.Set("access_token", p.token)

	url := fmt.Sprintf("https://api.particle.io/v1/devices/%s/ping", deviceId)
	url += "?" + queryParams.Encode()

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return false, fmt.Errorf("particle.Ping: http.NewRequest: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("particle.Ping: client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("particle.Ping: io.ReadAll: %w", err)
	}

	// TODO: handle device offline error code as none error
	if resp.StatusCode != 200 {
		// This isnt really any error
		return false, fmt.Errorf("particle.Ping: status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	type ResponseData struct {
		Online bool `json:"online"`
		Ok     bool `json:"ok"`
	}
	var response ResponseData

	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, fmt.Errorf("particle.Ping: json.Unmarshal: %w", err)
	}
	return response.Online, nil
}

func (p Particle) CloudFunction(deviceId string, cloudFunction string, argument string, returnValue *int) (bool, error) {
	params := url.Values{}
	params.Add("access_token", p.token)
	params.Add("arg", argument)

	url := fmt.Sprintf("https://api.particle.io/v1/devices/%s/%s", deviceId, cloudFunction)

	// This can block for a long time
	resp, err := http.PostForm(url, params)
	if err != nil {
		return false, fmt.Errorf("particle.CloudFunction: http.NewRequest: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("particle.CloudFunction: io.ReadAll: %w,  body %s", err, body)
	}

	// TODO: Find a way to treat this differently?
	if resp.StatusCode != 200 {
		// time out response status code: 400, response body: {"ok":false,"error":"Timed out."}
		return false, fmt.Errorf("particle.CloudFunction: status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	type ResponseData struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Connected   bool   `json:"Connected"`
		ReturnValue int    `json:"return_value"`
	}
	var data ResponseData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return false, fmt.Errorf("particle.Ping: json.Unmarshal: %w, body %s", err, body)
	}

	return true, nil
}

// Makes a test request to particle to see if the token is valid, should get a 200 on a list device request
func testToken(token string) (bool, error) {
	params := url.Values{}
	params.Add("access_token", token)
	url := "https://api.particle.io/v1/devices"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("particle.TestToken: http.NewRequest: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("particle.TestToken: client.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 401 { // Bad token
		return false, nil
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("particle.TestToken: io.ReadAll: %w,  body %s", err, body)
		}
		return false, fmt.Errorf("particle.TestToken: Unexpected response from Particle: %d:, body: %s", resp.StatusCode, string(body))
	}
}
