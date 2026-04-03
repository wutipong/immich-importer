package immich

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Post[R any](
	ctx context.Context, server ServerConfig, path string, data any,
) (result R, err error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	body := bytes.NewBuffer(jsonData)

	return DoRequestWithReturnObject[R](
		ctx, "POST", server, path, body, "application/json",
	)
}

func DoRequestWithReturnObject[R any](
	ctx context.Context,
	method string,
	server ServerConfig,
	path string,
	body io.Reader,
	contentType string,
) (result R, err error) {
	resp, err := DoRequest(ctx, method, server, path, body, contentType)
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
	ctx context.Context,
	method string,
	server ServerConfig,
	path string,
	body io.Reader,
	contentType string,
) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(
		ctx, method, server.URL.JoinPath(path).String(), body,
	)
	if err != nil {
		return
	}
	req.Header.Set("x-api-key", server.APIKey)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return http.DefaultClient.Do(req)
}

func Get[R any](
	ctx context.Context, server ServerConfig, path string,
) (result R, err error) {
	return DoRequestWithReturnObject[R](ctx, "GET", server, path, nil, "")
}

func Put[R any](
	ctx context.Context, server ServerConfig, path string, request any,
) (result R, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return
	}

	return DoRequestWithReturnObject[R](
		ctx,
		"PUT",
		server,
		path,
		bytes.NewReader(body),
		"application/json",
	)
}
