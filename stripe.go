package main

import (
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

// ID of the product in a bundle discount
var HOODIE_IDS = [...]string{"prod_O2mPalbP4HvHJc", "prod_O2mMZ7N70DGrap", "prod_O2mMUWKTwsVC1U"}

type item struct {
	ID          string   `json:"id"`
	NAME        string   `json:"name"`
	PRICE       int64    `json:"price"`
	IMAGES      []string `json:"images"`
	DESCRIPTION string   `json:"description"`
}

func stripeGetProducts() []item {
	items := []item{}

	params := &stripe.ProductListParams{}
	params.Limit = stripe.Int64(100)
	// Only return a single page of results. This is useful for testing.
	params.Single = true
	i := product.List(params)

	for i.Next() {
		p := i.Product()

		// Trolled if no default price
		value := int64(0)

		if p.DefaultPrice != nil {
			price, _ := price.Get(p.DefaultPrice.ID, nil)
			value = price.UnitAmount
		}

		items = append(items, item{
			ID:          p.ID,
			NAME:        p.Name,
			DESCRIPTION: p.Description,
			IMAGES:      p.Images,
			PRICE:       value,
		})
	}

	return items
}

type intent struct {
	CLIENT_SECRET string `json:"clientSecret"`
}

// Calculate the order total on the server to prevent
// people from directly manipulating the amount on the client
// ! Optional TODO - optimise by only requesting the ID's we actually need from Stripe.
func calculateOrderAmount(cart_items []cart_item) int64 {
	all_items := stripeGetProducts()

	var total_price int64 = 0
	var total_items int = 0
	var maybe_discount bool = false

	for _, cart_item := range cart_items {
		for _, s := range HOODIE_IDS {
			fmt.Printf("This hoodie ID was checked: %v", s)
			if cart_item.ID == s {
				maybe_discount = true
				break
			}
		}
		price := findItemPrice(all_items, cart_item.ID)
		total_price += price
		total_items += 1
	}

	if !maybe_discount {
		return total_price
	}

	if total_items >= 3 {
		total_price -= 1000
	} else if total_items >= 2 {
		total_price -= 500
	}

	fmt.Printf("Total items was: %v", total_items)
	fmt.Printf("The total price was: %v\n", total_price)
	fmt.Printf("There was a discount potentially applied: %v", maybe_discount)

	// Price must be at least $0.50 AUD, as per Stripe's minimum
	return max(50, total_price)
}

func findItemPrice(items []item, id string) int64 {
	for i := 0; i < len(items); i++ {
		if items[i].ID == id {
			return items[i].PRICE
		}
	}
	return 0
}

func stripeCreatePaymentIntent(items []cart_item) intent {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(calculateOrderAmount(items)),
		Currency: stripe.String(string(stripe.CurrencyAUD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, _ := paymentintent.New(params)

	return intent{CLIENT_SECRET: pi.ClientSecret}
}
