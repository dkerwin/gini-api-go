package giniapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Timing struct
type Timing struct {
	Upload     time.Duration
	Processing time.Duration
}

// Calculate total time from upload + processing
func (t *Timing) Total() time.Duration {
	return t.Upload + t.Processing
}

// Page struct
type Page struct {
	Images     map[string]string `json:"images"`
	PageNumber int               `json:"pageNumber"`
}

// Links struct
type Links struct {
	Document    string `json:"document"`
	Extractions string `json:"extractions"`
	Layout      string `json:"layout"`
	Processed   string `json:"processed"`
}

// Document struct
type Document struct {
	Timing
	Client               *APIClient
	Owner                string
	Links                Links  `json:"_links"`
	CreationDate         int    `json:"creationDate"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Origin               string `json:"origin"`
	PageCount            int    `json:"pageCount"`
	Pages                []Page `json:"pages"`
	Progress             string `json:"progress"`
	SourceClassification string `json:"sourceClassification"`
}

// DocumentSet list of documents
type DocumentSet struct {
	TotalCount int         `json:"totalCount"`
	Documents  []*Document `json:"documents"`
}

// String representaion of a document
func (d *Document) String() string {
	return fmt.Sprintf(d.ID)
}

// poll state and return true when done
func (d *Document) Poll(timeout time.Duration) error {
	start := time.Now()
	defer func() { d.Timing.Processing = time.Since(start) }()

	respChannel := make(chan bool, 1)

	go func() {
		respChannel <- d.WaitForCompletion()
	}()

	select {
	case resp := <-respChannel:
		if resp == true {
			return nil
		}
	case <-time.After(timeout):
		return fmt.Errorf("Processing timeout after %v seconds", timeout.Seconds())
	}

	return nil
}

// Update document struct (self)
func (d *Document) Update() *Document {
	newDoc, _ := d.Client.Get(d.Links.Document, d.Owner)
	return newDoc
}

func (d *Document) WaitForCompletion() bool {
	for {
		doc, _ := d.Client.Get(d.Links.Document, d.Owner)
		if doc.Progress == "COMPLETED" || doc.Progress == "ERROR" {
			return true
		}
	}
	return false
}

// Delete method
func (d *Document) Delete() error {
	resp, err := d.Client.MakeAPIRequest("DELETE", d.Links.Document, nil, nil, "")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Failed to delete document %s: HTTP status %d", d.ID, resp.StatusCode)
	}

	return nil
}

// Report bug
func (d *Document) ErrorReport(summary string, description string) error {
	resp, err := d.Client.MakeAPIRequest("POST",
		fmt.Sprintf("%s/errorreport?summary=%s&description=%s",
			d.Links.Document,
			summary,
			description,
		), nil, nil, "")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to submit error report for document %s: HTTP status %d", d.ID, resp.StatusCode)
	}

	return err
}

// Layout
func (d *Document) GetLayout() (*Layout, error) {
	var layout Layout

	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Layout, nil, nil, "")

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get layout for document %s: HTTP status %d", d.ID, resp.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&layout)

	return &layout, err
}

// Extractions
func (d *Document) GetExtractions() (*Extractions, error) {
	var extractions Extractions

	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Extractions, nil, nil, "")

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get extractions for document %s: HTTP status %d", d.ID, resp.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&extractions)

	return &extractions, err
}

// Processed Document
func (d *Document) GetProcessed() ([]byte, error) {
	headers := map[string]string{
		"Accept": "application/octet-stream",
	}
	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Processed, nil, headers, "")

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Failed to get processed document %s: HTTP status %d", d.ID, resp.StatusCode)
	}

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	return buf.Bytes(), err
}

// SubmitFeedback on a single label
// func (d *Document) SubmitFeedback(key string, newValue string) error {

// 	if val, ok := e.Extractions[key]; ok {
// 		return val.Value
// 	}
// 	return ""
// }
