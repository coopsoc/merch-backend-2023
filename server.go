package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
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
	router.GET("/products", getProducts)

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

	c.IndentedJSON(http.StatusOK, items)
}
