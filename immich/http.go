package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func Post[R any](url *url.URL, data any, apiKey string) (result R, err error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	body := bytes.NewBuffer(jsonData)

	return DoRequestWithResult[R]("POST", url, body, "application/json", apiKey)
}

func DoRequestWithResult[R any](
	method string,
	url *url.URL,
	body io.Reader,
	contentType string,
	apiKey string,
) (result R, err error) {
	resp, err := DoRequest(method, url, body, contentType, apiKey)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBuff, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		err = fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBuff))
		return
	}

	err = json.Unmarshal(respBuff, &result)
	if err != nil {
		return
	}
	return
}

func DoRequest(
	method string,
	url *url.URL,
	body io.Reader,
	contentType string,
	apiKey string,
) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return
	}
	req.Header.Set("x-api-key", apiKey)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return http.DefaultClient.Do(req)
}

func Get[R any](url *url.URL, apiKey string) (result R, err error) {
	return DoRequestWithResult[R]("GET", url, nil, "", apiKey)
}

func Put[R any](url *url.URL, request any, apiKey string) (result R, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return
	}

	return DoRequestWithResult[R]("PUT", url, bytes.NewReader(body), "application/json", apiKey)
}
