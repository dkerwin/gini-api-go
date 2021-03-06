package giniapi

import (
	"context"
	"testing"
)

func Test_ExtractionsGetValue(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/extractions",
		},
	}

	ctx := context.Background()

	extractions, _ := doc.GetExtractions(ctx, false)
	assertEqual(t, extractions.GetValue("amountToPay"), "24.99:EUR", "")
	assertEqual(t, extractions.GetValue("unknown"), "", "")
}
