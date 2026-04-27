package httprequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func MakeRequest(ctx context.Context, method string, url string, body any) (any, error) {
	return MakeRequestWithHeaders(ctx, method, url, body, nil)
}

func MakeRequestWithHeaders(ctx context.Context, method string, url string, body any, headers map[string]string) (any, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		fmt.Println("error occured when trying to create a new request")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	

	resp,err := client.Do(req)
	if err != nil {
		fmt.Println("error occured when trying to make a request")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Something went wrong")
		return nil, fmt.Errorf("Something went wrong: %d", resp.StatusCode)
	}
     
	var data any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("error occured when trying to decode the response")
		return nil, err
	}

	return data, nil
}