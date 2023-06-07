# merch-backend

A backend (written in Go) used to process the purchase, order, delivery, and invoicing of the merchandise at https://coopsoc.com.au/merch.

The website will send various API requests to this backend (hosted locally on a private server), which will then use:

- The Stripe API to handle payments
- The Google Sheets API to store records of purchases
- A Go library to generate invoices

## Usage

Either:

```sh
go run .
```

or:

```sh
go build
./coopsoc.com.au
```

Might also need to run `go mod tidy`.

## Routes

Server will run on localhost:8080 by default.
Available routes include:

- GET: /products
- POST: /payment

To test the server is running, navigate to http://localhost:8080/products in a web browser.
