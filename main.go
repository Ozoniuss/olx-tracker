package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL   = "https://www.olx.ro/api/partner"
	apiVer    = "2.0"
	userAgent = "price-tracker/0.1"
)

type Client struct {
	HTTP        *http.Client
	AccessToken string
}

func NewClient(token string) *Client {
	return &Client{
		AccessToken: token,
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type AdvertResponse struct {
	Data Advert `json:"data"`
}

type Advert struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at"`
	ActivatedAt string `json:"activated_at"`
	ValidTo     string `json:"valid_to"`
	CategoryID  int    `json:"category_id"`

	Price *Price `json:"price,omitempty"`
}

type Price struct {
	Value      float64 `json:"value"`
	Currency   string  `json:"currency"`
	Negotiable bool    `json:"negotiable"`
}

func (c *Client) GetAdvert(ctx context.Context, advertID int) (*Advert, error) {
	url := fmt.Sprintf("%s/adverts/%d", baseURL, advertID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Version", apiVer)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("olx api error: status %d", resp.StatusCode)
	}

	var parsed AdvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return &parsed.Data, nil
}

func main() {
	// still sample. test HTTPS in container
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com")
	if err != nil {
		panic(err) // TLS / cert errors usually show up here
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body length:", len(body))

	fmt.Println("body: ", string(body))
}
