package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	router := gin.Default()

	// Maybe enable debug mode
	if os.Getenv("DEBUG") == "true" {
		gin.SetMode(gin.DebugMode)
	}

	router.GET("/products", getProducts)
	router.POST("/payment", createPaymentIntent)

	router.Run("localhost:8080")
}

type item struct {
	ID          string   `json:"id"`
	NAME        string   `json:"name"`
	PRICE       int64    `json:"price"`
	IMAGES      []string `json:"images"`
	DESCRIPTION string   `json:"description"`
}

func getProducts(c *gin.Context) {
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

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, items)
	} else {
		c.JSON(http.StatusOK, items)
	}
}

type cart_item struct {
	id string
}

type intent struct {
	CLIENT_SECRET string `json:"clientSecret"`
}

func calculateOrderAmount(items []cart_item) int64 {
	// Replace this constant with a calculation of the order's amount
	// Calculate the order total on the server to prevent
	// people from directly manipulating the amount on the client
	return 1400
}

func createPaymentIntent(c *gin.Context) {
	type paymentIntent struct {
		Items []cart_item `json:"items"`
	}

	var body paymentIntent
	c.BindJSON(&body)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(calculateOrderAmount(body.Items)),
		Currency: stripe.String(string(stripe.CurrencyAUD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, _ := paymentintent.New(params)

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, intent{CLIENT_SECRET: pi.ClientSecret})
	} else {
		c.JSON(http.StatusOK, intent{CLIENT_SECRET: pi.ClientSecret})
	}
}
