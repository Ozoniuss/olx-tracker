package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var ErrAdDeactivated = errors.New("olx ad was deactivated")

func FetchProduct(
	ctx context.Context,
	client *http.Client,
	url string,
) (*Product, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// OLX sends 410 for deactivated ads
	if resp.StatusCode == http.StatusGone {
		return nil, ErrAdDeactivated
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request responded with status %d", resp.StatusCode)
	}

	z := html.NewTokenizer(resp.Body)
	for {
		switch z.Next() {
		case html.ErrorToken:
			// some AD urls may return a 200 even though the page never existed
			// and eventually OLX will redirect to a 404 page. But it's written
			// as a garbage SPA and this redirect is implemented with javascript.
			// Moreover it will return different error messages depending on
			// the URL (IDxxx fails differently than simple xxxx)
			return nil, errors.New("product json-ld not found")

		case html.StartTagToken:
			t := z.Token()
			if t.Data == "script" && hasJSONLDType(t.Attr) {
				if z.Next() == html.TextToken {
					raw := strings.TrimSpace(z.Token().Data)
					return parseJSONLD(raw)
				}
			}
		}
	}
}

func hasJSONLDType(attrs []html.Attribute) bool {
	for _, a := range attrs {
		if a.Key == "type" && a.Val == "application/ld+json" {
			return true
		}
	}
	return false
}

func parseJSONLD(data string) (*Product, error) {
	var productInfo Product
	if err := json.Unmarshal([]byte(data), &productInfo); err != nil {
		log.Println("Invalid JSON-LD:", err)
		return nil, fmt.Errorf("failed to parse product info: %w", err)
	}

	return &productInfo, nil
}
