package giniapi

import (
	"fmt"
)

// Very simplistic example. You shoud have a lot more error handling in place
func Example_typicalOauth2Flow() {
	// Setup api connection
	api, err := giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Username:       "user1",
		Password:       "secret",
		Authentication: "oauth2",
	})

	if err != nil {
		panic("Gini API login failed: %s", err)
	}

	// Read a PDF document
	bodyBuf, _ := os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ := api.Upload(bodyBuf, "invoice.pdf", "", "")

	// Get extractions from our uploaded document
	extractions, _ := doc.GetExtractions()

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))
	// Output: IBAN has been found: DE2937030181004142612 Woohoo!
}

// Very simplistic example. You shoud have a lot more error handling in place
func Example_typicalBasicAuthFlow() {
	// Setup api connection
	api, err := giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Authentication: "enterprise",
	})

	if err != nil {
		panic("Gini API login failed: %s", err)
	}

	// Read a PDF document
	bodyBuf, _ := os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ := api.Upload(bodyBuf, "invoice.pdf", "", "user123")

	// Get extractions from our uploaded document
	extractions, _ := doc.GetExtractions()

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))
	// Output: IBAN has been found: DE2937030181004142612 Woohoo!
}
