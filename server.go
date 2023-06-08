// TODO - run append (consumer & product) / update functions in sheets.go on PaymentIntent creation & webhook (respectively)

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
)

func main() {
	// gmain()

	godotenv.Load()

	ok := false
	stripe.Key, ok = os.LookupEnv("STRIPE_KEY")
	if !ok {
		log.Fatal("STRIPE_KEY not set")
	}

	// Maybe enable debug mode
	if os.Getenv("DEBUG") == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(CORSMiddleware())

	router.GET("/products", getProducts)
	router.POST("/payment", createPaymentIntent)

	port := os.Getenv("PORT")
	router.Run(":" + port)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getProducts(c *gin.Context) {
	items := stripeGetProducts()

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

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, i)
	} else {
		c.JSON(http.StatusOK, i)
	}
}
