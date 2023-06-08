// TODO - run append (consumer & product) / update functions in sheets.go on PaymentIntent creation & webhook (respectively)

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"

	"github.com/google/uuid"
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

type checkout_info struct {
	FIRST_NAME string      `json:"first_name"`
	LAST_NAME  string      `json:"last_name"`
	EMAIL      string      `json:"email"`
	CART_ITEMS []cart_item `json:"cart_items"`
}

func createPaymentIntent(c *gin.Context) {
	var body checkout_info
	c.BindJSON(&body)
	// TODO - make sure data is passed in correctly

	// Stripe
	i := stripeCreatePaymentIntent(body.CART_ITEMS)

	// Google sheets API
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	guid := uuid.New().String()

	consumer := Consumer{
		GUID:      guid,
		FirstName: body.FIRST_NAME,
		LastName:  body.LAST_NAME,
		Email:     body.EMAIL,
	}

	appendUserInfo(spreadsheetID, consumer)

	for _, cart_item := range body.CART_ITEMS {
		// TODO - create lookup function to return product name for given cart_item.id?
		product := Product{
			GUID:          guid,
			ProductName:   cart_item.id,
			ProductColour: cart_item.color,
			ProductSize:   cart_item.size,
			PaymentStatus: "", // Will get updated by the webhook later
		}

		appendProductInfo(spreadsheetID, product)
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// if dev, use indented json
	if gin.Mode() == gin.DebugMode {
		c.IndentedJSON(http.StatusOK, i)
	} else {
		c.JSON(http.StatusOK, i)
	}
}
