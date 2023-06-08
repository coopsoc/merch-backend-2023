package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Struct to contain all of the various field that are needed to append into the spreadsheet
// Sheet 1
type Consumer struct { // Corresponding column:
	GUID      string // 	A
	FirstName string // 	B
	LastName  string // 	C
	Email     string // 	D
}

// Struct to contain all of the various field for the product
// Sheet 2
type Product struct { // Corresponding column:
	GUID          string //	A
	ProductName   string //	B
	ProductColour string //	C
	ProductSize   string //	D
	PaymentStatus string //	E
}

// Function that will append the user data into the spreadsheet
// TODO - don't append user info if GUID already exists? (Or perhaps update fields)
func appendUserInfo(spreadsheetID string, consumer Consumer) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Specifying the writing range
	cellRange := "Sheet1!A2:D"

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

	appendResp, err := srv.Spreadsheets.Values.Append(spreadsheetID, cellRange, valueRange).ValueInputOption(valueInputOption).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	fmt.Printf("Updated range: %s\n", appendResp.Updates.UpdatedRange)

	return nil
}

// Function that will append the product data into the spreadsheet
func appendProductInfo(spreadsheetID string, product Product) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Specifying the writing range
	cellRange := "Sheet2!A1:D"

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

	appendResp, err := srv.Spreadsheets.Values.Append(spreadsheetID, cellRange, valueRange).ValueInputOption(valueInputOption).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	fmt.Printf("Updated range: %s\n", appendResp.Updates.UpdatedRange)

	return nil
}

// Optional todo - rewrite to filter data by the GUID column (find row w/ corresponding GUID,
// then update corresponding status column)

// Function that will update the row status to either fail or success
func orderStatusUpdate(spreadsheetID string, GUID string, PaymentStatus string) error {
	ctx := context.Background()

	srv, err := getSheetsClient(ctx)
	if err != nil {
		return err
	}

	// Define the range to search in, including the GUID and status columns
	searchRange := "Sheet2"

	// Create the request to search for the GUID value
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, searchRange).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("no data found")
	}

	// Iterate over the rows to find the matching GUID and check the status
	for i, row := range resp.Values {
		if row[0] == GUID {
			// Write PaymentStatus to column E, row (i + 1) (needs to be 1-indexed)

			// Specifying the writing range
			cellRange := fmt.Sprint("Sheet2!E", i+1, ":E") // Range in the format "Sheet2!E1:E"

			// Prepare the value to update column E (PaymentStatus) to
			values := []interface{}{
				PaymentStatus,
			}

			// Specifying the ranges of values that will be appended to the Google Sheet
			// AKA fields of the request body
			valueRange := &sheets.ValueRange{
				Values: [][]interface{}{values},
			}

			// Indicating that values entered should be treated as if users themselves had entered them
			valueInputOption := "USER_ENTERED"

			updateResp, err := srv.Spreadsheets.Values.Update(spreadsheetID, cellRange, valueRange).ValueInputOption(valueInputOption).Context(ctx).Do()

			if err != nil {
				return fmt.Errorf("unable to update PaymentStatus: %v", err)
			}

			fmt.Printf("Updated range: %s\n", updateResp.UpdatedRange)

			break
		}
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
