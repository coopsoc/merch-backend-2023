// TODO - run append (consumer & product) / update functions in sheets.go on PaymentIntent creation & webhook (respectively)

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
)

func main() {

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
	router.Any("/webhook", updatePaymentStatus)

	port := os.Getenv("PORT")
	router.Run(":" + port)

	fmt.Printf("Server is running...\n")
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

func updatePaymentStatus(c *gin.Context) {
	var event stripe.Event
	c.BindJSON(&event)

	switch event.Type {
	case "payment_intent.succeeded":
		var intent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &intent)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		updateOrderStatus(SPREADSHEET_ID, intent.ClientSecret, "Approved")
	case "payment_intent.failed":
		log.Println("Payment failed")
	}

	c.Status(http.StatusOK)
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

type cart_item struct {
	ID   string `json:"id" binding:"required"`
	SIZE string `json:"size" binding:"required"`
}

func createPaymentIntent(c *gin.Context) {
	type cart struct {
		ITEMS []cart_item `json:"items"`
	}

	type request struct {
		FIRST_NAME string `json:"firstName"`
		LAST_NAME  string `json:"lastName"`
		EMAIL      string `json:"email"`
		CART       cart   `json:"cart" binding:"required"`
	}

	var body request
	c.BindJSON(&body)

	i := stripeCreatePaymentIntent(body.CART.ITEMS, body.EMAIL)

	var consumer Consumer

	consumer.FirstName = body.FIRST_NAME
	consumer.LastName = body.LAST_NAME
	consumer.Email = body.EMAIL

	appendUserInfo(SPREADSHEET_ID, consumer)

	for _, v := range body.CART.ITEMS {
		var product Product
		product.ClientSecret = i.CLIENT_SECRET
		product.FirstName = body.FIRST_NAME
		product.LastName = body.LAST_NAME
		product.ProductName = filter(stripeGetProducts(), func(i item) bool {
			return i.ID == v.ID
		})
		product.ProductSize = v.SIZE
		product.PaymentStatus = "Unapproved"

		// this should not happen here
		appendProductInfo(SPREADSHEET_ID, product)
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, i)
	} else {
		c.JSON(http.StatusOK, i)
	}
}
