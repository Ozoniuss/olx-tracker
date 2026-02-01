package product

import "fmt"

type Product struct {
	Context     string   `json:"@context"`
	Type        string   `json:"@type"`
	Name        string   `json:"name"`
	Image       []string `json:"image"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	SKU         string   `json:"sku"`
	Offers      Offer    `json:"offers"`
}

type Offer struct {
	Type          string   `json:"@type"`
	Availability  string   `json:"availability"`
	AreaServed    Area     `json:"areaServed"`
	PriceCurrency string   `json:"priceCurrency"`
	Price         float64  `json:"price"`
	Shipping      Shipping `json:"shippingDetails"`
	ItemCondition string   `json:"itemCondition"`
}

type Area struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type Shipping struct {
	Type                string         `json:"@type"`
	ShippingRate        MonetaryAmount `json:"shippingRate"`
	ShippingDestination Region         `json:"shippingDestination"`
}

type MonetaryAmount struct {
	Type     string `json:"@type"`
	Currency string `json:"currency"`
}

type Region struct {
	Type           string `json:"@type"`
	AddressCountry string `json:"addressCountry"`
}

func PrintRelevantProductInfo(p *Product) {
	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Description: %s\n", p.Description)
	fmt.Printf("URL: %s\n", p.URL)
	fmt.Printf("Price: %.2f %s\n", p.Offers.Price, p.Offers.PriceCurrency)
	fmt.Printf("Availability: %s\n", p.Offers.Availability)
	fmt.Printf("Condition: %s\n", p.Offers.ItemCondition)
}
