package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
)

func main() {
	stripe.Key = os.Getenv("STRIPE_KEY")

	// Maybe enable debug mode
	if os.Getenv("DEBUG") == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/products", getProducts)
	router.POST("/payment", createPaymentIntent)

	port := os.Getenv("PORT")
	router.Run(":" + port)
}

func getProducts(c *gin.Context) {
	items := stripeGetProducts()

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, items)
	} else {
		c.JSON(http.StatusOK, items)
	}
}

func createPaymentIntent(c *gin.Context) {
	type cart struct {
		Items []cart_item `json:"items"`
	}

	var body cart
	c.BindJSON(&body)
	// TODO - make sure cart items have quantities and IDs

	i := stripeCreatePaymentIntent(body.Items)

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, i)
	} else {
		c.JSON(http.StatusOK, i)
	}
}
