package httprequest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func MakeRequest(ctx context.Context,method string, url string) (any, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

    // ctx,cancel := context.WithTimeout(context.Background(), time.Second * 10)
	// defer cancel()

	req, err := http.NewRequestWithContext(ctx,method,url,nil)
	if err != nil {
		fmt.Println("error occured when trying to create a new request")
		return nil, err
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