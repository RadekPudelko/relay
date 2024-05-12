package particle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// https://docs.particle.io/reference/cloud-apis/api/#errors
func Ping(somId string, productId int, token string) (bool, error) {
    queryParams := url.Values{}
    queryParams.Set("access_token", token)

    url := fmt.Sprintf("https://api.particle.io/v1/products/%d/devices/%s/ping", productId, somId)
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
    // TODO: According to particle a 408 should be returned on timeout
    if resp.StatusCode != 200 {
        // This isnt really any error
        return false, fmt.Errorf("particle.Ping: status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    type ResponseData struct {
        Online bool `json:"online"`
        Ok bool `json:"ok"`
    }
    var response ResponseData

    err = json.Unmarshal(body, &response)
    if err != nil {
        return false, fmt.Errorf("particle.Ping: json.Unmarshal: %w", err)
    }
    return response.Online, nil
}

func CloudFunction(somId string, productId int, cloudFunction string, argument string, token string) (int, error) {
    params := url.Values{}
    params.Add("access_token", token)
    params.Add("arg", argument)

    url := fmt.Sprintf("https://api.particle.io/v1/products/%d/devices/%s/%s", 
        productId, somId, cloudFunction)

    // This can block for a long time
    resp, err := http.PostForm(url, params)
    if err != nil {
        return 0, fmt.Errorf("particle.CloudFunction: http.NewRequest: %w", err)
    }
	defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
	if err != nil {
        return 0, fmt.Errorf("particle.CloudFunction: io.ReadAll: %w, body %s", err, body)
	}

    // TODO: Find a way to treat this differently?
    if resp.StatusCode != 200 {
        return 0, fmt.Errorf("particle.CloudFunction: status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    type ResponseData struct {
        Id string `json:"id"`
        Name string `json:"name"`
        Connected bool `json:"Connected"`
        ReturnValue int `json:"return_value"`
    }
    var data ResponseData

    err = json.Unmarshal(body, &data)
    if err != nil {
        return 0, fmt.Errorf("particle.Ping: json.Unmarshal: %w, body %s", err, body)
    }

    return data.ReturnValue, nil
}

