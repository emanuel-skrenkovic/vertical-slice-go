package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type responseAssertion func(*http.Response)

func sendRequest[TReq any, TResp any](
	c *http.Client,
	url string,
	method string,
	req TReq,
	opts ...responseAssertion,
) (TResp, error) {
	var resp TResp

	payload, err := json.Marshal(req)
	if err != nil {
		return resp, err
	}

	httpReq, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return resp, err
	}

	httpResp, err := c.Do(httpReq)
	if err != nil {
		return resp, err
	}

	for _, opt := range opts {
		opt(httpResp)
	}

	if httpResp.ContentLength > 0 {
		defer func() {
			_ = httpResp.Body.Close()
		}()

		responsePayload, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return resp, err
		}

		if err := json.Unmarshal(responsePayload, &resp); err != nil {
			return resp, err
		}
	}

	return resp, err
}
