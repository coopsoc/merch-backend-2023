# merch-backend
A backend (written in Go) used to process the purchase, order, delivery, and invoicing of the merchandise at https://coopsoc.com.au/merch.

The website will send various API requests to this backend (hosted locally on a private server), which will then use:
- The Stripe API to handle payments/refunds
- The Google Sheets API to store records of purchases
- A Go library to generate invoices
