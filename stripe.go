package main

import (
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

// ID of the product in a bundle discount
var HOODIE_IDS = [...]string{
	"prod_O2mPalbP4HvHJc", // Hoodie - Cream
	"prod_O2mMZ7N70DGrap", // Hoodie - Green
	"prod_O2mMUWKTwsVC1U", // Hoodie - Black
	"prod_O34lEdsgMSE8TY", // Hoodie Cream - Stealth
	"prod_O34mEl0jE7T9UJ", // Hoodie Green - Stealth
	"prod_O34lSQer1T3zUX", // Hoodie Black - Stealth
}

const HOODIE_AND_ONE_ITEM_DISCOUNT = 500
const HOODIE_AND_TWO_ITEMS_DISCOUNT = 1000

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
	var hoodie_in_cart bool = false

	for _, cart_item := range cart_items {
		hoodie_in_cart = hoodie_in_cart || itemIsHoodie(cart_item.ID)
		total_price += findItemPrice(all_items, cart_item.ID)
		total_items++
	}

	if !hoodie_in_cart {
		fmt.Print("\tNo hoodies in cart. Discount not applied.\n")
		return total_price
	}

	if total_items >= 3 {
		total_price -= HOODIE_AND_TWO_ITEMS_DISCOUNT
	} else if total_items >= 2 {
		total_price -= HOODIE_AND_ONE_ITEM_DISCOUNT
	}

	fmt.Printf("\tTotal items was: %v\n", total_items)
	fmt.Printf("\tThe total price was: %v\n", total_price)
	fmt.Printf("\tDiscount applied: %v\n", hoodie_in_cart)

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

func itemIsHoodie(item_id string) bool {
	for _, hoodie_id := range HOODIE_IDS {
		fmt.Printf("\tThis hoodie ID was checked: %v\n", hoodie_id)
		if item_id == hoodie_id {
			fmt.Printf("\tThis hoodie ID was matched: %v\n", hoodie_id)
			return true
		}
	}
	fmt.Printf("\tThis item is not a hoodie: %v\n", item_id)
	return false
}

func stripeCreatePaymentIntent(items []cart_item, email string) intent {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(calculateOrderAmount(items)),
		Currency: stripe.String(string(stripe.CurrencyAUD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		ReceiptEmail: stripe.String(email),
	}

	pi, _ := paymentintent.New(params)

	return intent{CLIENT_SECRET: pi.ClientSecret}
}
