package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Post[R any](server ServerConfig, path string, data any) (result R, err error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	body := bytes.NewBuffer(jsonData)

	return DoRequestWithReturnObject[R]("POST", server, path, body, "application/json")
}

func DoRequestWithReturnObject[R any](
	method string,
	server ServerConfig,
	path string,
	body io.Reader,
	contentType string,
) (result R, err error) {
	resp, err := DoRequest(method, server, path, body, contentType)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBuff, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode >= 300 {
		err = fmt.Errorf(
			"request failed with status code %d: %s",
			resp.StatusCode,
			string(respBuff),
		)
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
	server ServerConfig,
	path string,
	body io.Reader,
	contentType string,
) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, server.URL.JoinPath(path).String(), body)
	if err != nil {
		return
	}
	req.Header.Set("x-api-key", server.APIKey)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return http.DefaultClient.Do(req)
}

func Get[R any](server ServerConfig, path string) (result R, err error) {
	return DoRequestWithReturnObject[R]("GET", server, path, nil, "")
}

func Put[R any](server ServerConfig, path string, request any) (result R, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return
	}

	return DoRequestWithReturnObject[R](
		"PUT",
		server,
		path,
		bytes.NewReader(body),
		"application/json",
	)
}
