package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ozoniuss/olx-tracker/config"
	dbpkg "github.com/Ozoniuss/olx-tracker/internal/db"
	productpkg "github.com/Ozoniuss/olx-tracker/internal/product"
)

func main() {

	debug()
	os.Exit(0)

	c, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(c)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	ctx := context.Background()

	db, err := dbpkg.ConnectToPostgres(
		ctx,
		dbpkg.GetPostgresURL(c.Postgres),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	uid, err := dbpkg.GetUserID(ctx, db, "testuser", "testuser")
	if err != nil {
		if errors.Is(err, dbpkg.ErrNotFound) {
			uid, err = dbpkg.NewUser(ctx, db, "testuser", "testuser", false)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	fmt.Printf("User ID: %s\n", uid)

	ads := []string{
		`https://www.olx.ro/d/oferta/mouse-gaming-logitech-pro-x-superlight-IDkbEDA.html`,
		`https://www.olx.ro/d/oferta/motocicleta-ktm-790-duke-24-IDgV8Gy.html`,
		`https://www.olx.ro/d/oferta/inchiriere-apartament-2-camere-IDkdRuG.html`,
	}

	// track the adds if they aren't there
	for _, adurl := range ads {
		err = dbpkg.TrackAddForUser(ctx, db, uid, adurl)
		if err != nil {
			if errors.Is(err, dbpkg.ErrAlreadyExists) {
				fmt.Println("Product is already being tracked for this user.")
			}
		}

	}

	trackedProducts, err := dbpkg.ListTrackedProductsForUser(ctx, db, uid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tracked URLs for user %s: %v\n", uid, trackedProducts)

	for _, tp := range trackedProducts {
		product, err := productpkg.FetchProduct(
			ctx,
			client,
			tp.URL,
		)
		if err != nil {
			log.Fatal(err)
		}

		if tp.URL != product.URL {
			log.Fatalf("Fetched product URL (%s) does not match requested URL (%s)", product.URL, tp.URL)
		}

		productpkg.PrintRelevantProductInfo(product)

		rawjson, err := json.Marshal(product)
		if err != nil {
			log.Fatal(err)
		}

		dbpkg.StoreNextAddSnapshot(ctx,
			db,
			tp.ID,
			product.Name,
			product.Description,
			int64(product.Offers.Price*100),
			product.Offers.PriceCurrency,
			product.Offers.Availability,
			rawjson,
		)
		fmt.Println("Stored snapshot for", product.URL)
	}
}

func debug() {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	ctx := context.Background()

	ads := []string{
		// `https://www.olx.ro/d/oferta/mouse-gaming-logitech-pro-x-superlight-IDkbEDA.html`,
		// `https://www.olx.ro/d/oferta/motocicleta-ktm-790-duke-24-IDgV8Gy.html`,
		// `https://www.olx.ro/d/oferta/inchiriere-apartament-2-camere-IDkdRuG.html`,
		// `https://www.olx.ro/d/oferta/apple-magic-mouse-2-reincarcabil-port-lightning-sdasjasd.html`,
		`https://www.olx.ro/d/oferta/apple-magic-mouse-2-reincarcabil-port-lightning-IDjTmhe.html`,
	}

	for _, ad := range ads {
		product, err := productpkg.FetchProduct(
			ctx,
			client,
			ad,
		)
		if err != nil {
			log.Fatal(err)
		}
		productpkg.PrintRelevantProductInfo(product)

	}
}
