package main

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

// ID of the product in a bundle discount
var HOODIE_IDS = [...]string{"prod_o1xqaojym2o4ur"}

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

type cart_item struct {
	id   string
	size string
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

	for i := 0; i < len(cart_items); i++ {
		for _, s := range HOODIE_IDS {
			if cart_items[i].id == s {
				maybe_discount = true
				break
			}
		}
		total_price += findItemPrice(all_items, cart_items[i].id)
	}

	if maybe_discount && total_items >= 3 {
		// $10 off if you buy a hoodie and 2 other items
		total_price -= 1000
	} else if maybe_discount && total_items >= 2 {
		// $5 off if you buy a hoodie and 1 other item
		total_price -= 500
	}

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
