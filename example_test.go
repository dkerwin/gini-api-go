package giniapi_test

import (
	"fmt"
	"github.com/dkerwin/gini-api-go"
	"log"
	"os"
)

// Very simplistic example. You shoud have a lot more error handling in place
func ExampleNewClient() {

	//////////////////////////////////
	// Oauth2
	//////////////////////////////////

	// Setup api connection
	api, err := giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Username:       "user1",
		Password:       "secret",
		Authentication: "oauth2",
	})

	if err != nil {
		log.Panicf("Gini API login failed: %s", err)
	}

	// Read a PDF document
	bodyBuf, _ := os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ := api.Upload(bodyBuf, "invoice.pdf", "", "", 10)

	// Get extractions from our uploaded document
	extractions, _ := doc.GetExtractions()

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))

	//////////////////////////////////
	// basic Auth
	//////////////////////////////////

	// Setup api connection
	api, err = giniapi.NewClient(&giniapi.Config{
		ClientID:       "MY_CLIENT_ID",
		ClientSecret:   "********",
		Authentication: "basicAuth",
	})

	if err != nil {
		log.Panicf("Gini API login failed: %s", err)
	}

	// Read a PDF document
	bodyBuf, _ = os.Open("/tmp/invoice.pdf")

	// Upload document to gini without doctype hint and user identifier
	doc, _ = api.Upload(bodyBuf, "invoice.pdf", "", "user123", 10)

	// Get extractions from our uploaded document
	extractions, _ = doc.GetExtractions()

	// Print IBAN
	fmt.Printf("IBAN has been found: %s Woohoo!\n", extractions.GetValue("iban"))
}
