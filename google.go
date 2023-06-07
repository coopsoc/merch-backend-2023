package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Struct to contain all of the various field that are needed to append into the spreadsheet
type Consumer struct {
	GUID      string
	FirstName string
	LastName  string
	Email     string
}

// Struct to contain all of the various field for the product
type Product struct {
	GUID          string
	ProductName   string
	ProductColour string
	ProductSize   string
	PaymentStatus string
}

func gmain() {
	// Setting the spreadsheetId
	spreadsheetId := "1EbJIzUrwXX0NMKPwg941sDyI1LSmZuNO_w7B3xO_Y6I"

	consumer := Consumer{
		GUID:      "CONSUMER_GUID",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	product := Product{
		GUID:          "Product A",
		ProductName:   "Medium",
		ProductColour: "Blue",
		ProductSize:   "M",
		PaymentStatus: "success",
	}

	err := appendUserInfo(spreadsheetId, consumer)

	if err != nil {
		log.Fatalf("Failed to append user information: %v", err)
	}

	err_2 := appendProductInfo(spreadsheetId, product)

	if err_2 != nil {
		log.Fatalf("Failed to append product information: %v", err_2)
	}

	err_3 := orderStatusUpdate(spreadsheetId, consumer.GUID)

	if err_3 != nil {
		log.Fatalf("Failed to update the status of the product or user information: %v", err_3)
	}
}

// Function that will append the user data into the spreadsheet
func appendUserInfo(spreadsheetID string, consumer Consumer) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Prepare the values to be appended
	values := []interface{}{
		consumer.GUID,
		consumer.FirstName,
		consumer.LastName,
		consumer.Email,
	}

	// Specifying the ranges of values that will be appended to the Google Sheet
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// Indicating that values entered should be treated as if users themselves had entered them
	valueInputOption := "USER_ENTERED"

	// Specifying the writing range
	writeRange := "Sheet1!A2:D"
	appendResp, err := srv.Spreadsheets.Values.Append(spreadsheetID, writeRange, valueRange).ValueInputOption(valueInputOption).Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	fmt.Printf("Updated range: %s\n", appendResp.Updates.UpdatedRange)

	return nil
}

// Function that will append the product data into the spreadsheet
func appendProductInfo(spreadsheetId string, product Product) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Prepare the values to be appended
	values := []interface{}{
		product.GUID,
		product.ProductName,
		product.ProductColour,
		product.ProductSize,
		product.PaymentStatus,
	}

	// Specifying the ranges of values that will be appended to the Google Sheet
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// Indicating that values entered should be treated as if users themselves had entered them
	valueInputOption := "USER_ENTERED"

	// Specifying the writing range
	writeRange := "Sheet2!A2:D"
	appendResp, err := srv.Spreadsheets.Values.Append("1EbJIzUrwXX0NMKPwg941sDyI1LSmZuNO_w7B3xO_Y6I", writeRange, valueRange).ValueInputOption(valueInputOption).Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	fmt.Printf("Updated range: %s\n", appendResp.Updates.UpdatedRange)

	return nil
}

// THIS FUNCTION IS INCOMPLETE WE STILL NEED TO FIGURE OUT HOW TO FILTER THE COLUMN BY THE GUID
// WE THINK IT IS THE FILTER .BATCHUPDATE FUNCTION IN THE SHEETS API BUT WE ARE UNSURE

// Function that will update the row status to either fail or success
func orderStatusUpdate(spreadsheetId string, guid string) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Define the range to search in, including the GUID and status columns
	searchRange := "Sheet2!E2:E"

	// Create the request to search for the GUID value
	resp, err := srv.Spreadsheets.Values.Get("1EbJIzUrwXX0NMKPwg941sDyI1LSmZuNO_w7B3xO_Y6I", searchRange).Do()

	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("no data found")
	}

	// Iterate over the rows to find the matching GUID and check the status
	for _, col := range resp.Values {
		// Print columns E, which correspond to index 0 of the specified searchRange 'Sheet2!E2:E'.
		fmt.Printf("%s\n", col[0])
	}
	return nil
}

// Function to retrieve the Google Sheets client
func getSheetsClient(ctx context.Context) (*sheets.Service, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account file to config: %v", err)
	}
	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return srv, nil
}
