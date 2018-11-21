package giniapi

import (
	"context"
	"testing"
	"time"
)

func Test_TimingTotal(t *testing.T) {
	timing := Timing{
		Upload:     2,
		Processing: 5,
	}

	assertEqual(t, timing.Total(), time.Duration(7), "")
}

func Test_DocumentString(t *testing.T) {
	doc := Document{
		ID: "fb9877fc-f23c-40df-9e81-26e51f26682d",
	}

	assertEqual(t, doc.String(), "fb9877fc-f23c-40df-9e81-26e51f26682d", "Document.String() should return document ID")
}

func Test_DocumentUpdate(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Name:   "original",
		Links: Links{
			Document: testHTTPServer.URL + "/test/document/update",
		},
	}

	ctx := context.Background()
	resp := doc.Update(ctx)

	assertEqual(t, resp.Error, nil, "")
	assertEqual(t, doc.Name, "Updated!", "")
}

func Test_DocumentDelete(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Document: testHTTPServer.URL + "/test/document/delete",
		},
	}

	ctx := context.Background()
	resp := doc.Delete(ctx)

	assertEqual(t, resp.Error, nil, "")
}

func Test_DocumentGetLayout(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Layout: testHTTPServer.URL + "/test/layout",
		},
	}

	ctx := context.Background()
	_, resp := doc.GetLayout(ctx)

	assertEqual(t, resp.Error, nil, "")
}

func Test_DocumentGetExtractions(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/extractions",
		},
	}

	ctx := context.Background()
	_, resp := doc.GetExtractions(ctx, false)

	assertEqual(t, resp.Error, nil, "")
}

func Test_DocumentGetProcessed(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Processed: testHTTPServer.URL + "/test/processed",
		},
	}

	ctx := context.Background()
	docBytes, resp := doc.GetProcessed(ctx)

	assertEqual(t, resp.Error, nil, "")
	assertEqual(t, string(docBytes), "get processed", "")
}

func Test_DocumentSubmitFeedback(t *testing.T) {
	doc := Document{
		client: testOauthClient(t),
		Links: Links{
			Extractions: testHTTPServer.URL + "/test/feedback",
		},
	}

	feedback := map[string]map[string]interface{}{
		"iban": map[string]interface{}{
			"entity": "iban",
			"value":  "DE22222111117777766666",
		},
	}

	ctx := context.Background()

	resp := doc.SubmitFeedback(ctx, feedback)

	// single label
	assertEqual(t, resp.Error, nil, "")

	feedback["bic"] = map[string]interface{}{
		"entity": "bic",
		"value":  "HYVEDEMMXXX",
	}

	resp = doc.SubmitFeedback(ctx, feedback)

	// multiple labels
	assertEqual(t, resp.Error, nil, "")
}
