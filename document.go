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

// Total returns the summarized timings of upload and processing
func (t *Timing) Total() time.Duration {
	return t.Upload + t.Processing
}

// Page describes a documents pages
type Page struct {
	Images     map[string]string `json:"images"`
	PageNumber int               `json:"pageNumber"`
}

// Links contains the links to a documents resources
type Links struct {
	Document    string `json:"document"`
	Extractions string `json:"extractions"`
	Layout      string `json:"layout"`
	Processed   string `json:"processed"`
}

// Document contains all informations about a single document
type Document struct {
	Timing
	client               *APIClient // client is not exported
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

// DocumentSet is a list of documents with the total count
type DocumentSet struct {
	TotalCount int         `json:"totalCount"`
	Documents  []*Document `json:"documents"`
}

// String representaion of a document
func (d *Document) String() string {
	return fmt.Sprintf(d.ID)
}

// Poll the progress state of a document and return nil when the processing
// has completed (successful or failed). On timeout return error
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
		return fmt.Errorf("processing timeout after %v seconds", timeout.Seconds())
	}

	return nil
}

// Update document struct from self-contained document link
func (d *Document) Update() error {
	newDoc, err := d.client.Get(d.Links.Document, d.Owner)

	if err == nil {
		d.Owner = newDoc.Owner
		d.Links = newDoc.Links
		d.CreationDate = newDoc.CreationDate
		d.ID = newDoc.ID
		d.Name = newDoc.Name
		d.Origin = newDoc.Origin
		d.PageCount = newDoc.PageCount
		d.Pages = newDoc.Pages
		d.Progress = newDoc.Progress
		d.SourceClassification = newDoc.SourceClassification
	}

	return err
}

// WaitForCompletion checks document progress and returns true on
// COMPLETED or ERROR
func (d *Document) WaitForCompletion() bool {
	for {
		doc, _ := d.client.Get(d.Links.Document, d.Owner)
		if doc.Progress == "COMPLETED" || doc.Progress == "ERROR" {
			return true
		}
	}
	return false
}

// Delete a document
func (d *Document) Delete() error {
	resp, err := d.client.MakeAPIRequest("DELETE", d.Links.Document, nil, nil, "")

	if err != nil {
		return err
	}

	return CheckHTTPStatus(resp.StatusCode, http.StatusNoContent,
		fmt.Sprintf("failed to delete document %s: HTTP status %d", d.ID, resp.StatusCode))
}

// ErrorReport creates a bug report in Gini's bugtracking system. It's a convinience way
// to help Gini learn from difficult documents
func (d *Document) ErrorReport(summary string, description string) error {
	resp, err := d.client.MakeAPIRequest("POST",
		fmt.Sprintf("%s/errorreport?summary=%s&description=%s",
			d.Links.Document,
			summary,
			description,
		), nil, nil, "")

	if err != nil {
		return err
	}

	return CheckHTTPStatus(resp.StatusCode, http.StatusOK,
		fmt.Sprintf("failed to submit error report for document %s: HTTP status %d", d.ID, resp.StatusCode))
}

// GetLayout returns the JSON representation of a documents layout parsed as
// Layout struct
func (d *Document) GetLayout() (*Layout, error) {
	var layout Layout

	resp, err := d.client.MakeAPIRequest("GET", d.Links.Layout, nil, nil, "")

	if err != nil {
		return nil, err
	}

	if err := CheckHTTPStatus(resp.StatusCode, http.StatusOK,
		fmt.Sprintf("failed to get layout for document %s: HTTP status %d", d.ID, resp.StatusCode)); err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&layout)

	return &layout, err
}

// GetExtractions returns a documents extractions in a Extractions struct
func (d *Document) GetExtractions() (*Extractions, error) {
	var extractions Extractions

	resp, err := d.client.MakeAPIRequest("GET", d.Links.Extractions, nil, nil, "")

	if err != nil {
		return nil, err
	}

	if err := CheckHTTPStatus(resp.StatusCode, http.StatusOK,
		fmt.Sprintf("failed to get extractions for document %s: HTTP status %d", d.ID, resp.StatusCode)); err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&extractions)

	return &extractions, err
}

// GetProcessed returns a byte array of the processed (rectified, optimized) document
func (d *Document) GetProcessed() ([]byte, error) {
	headers := map[string]string{
		"Accept": "application/octet-stream",
	}

	resp, err := d.client.MakeAPIRequest("GET", d.Links.Processed, nil, headers, "")

	if err != nil {
		return nil, err
	}

	if err := CheckHTTPStatus(resp.StatusCode, http.StatusOK,
		fmt.Sprintf("failed to get processed document %s: HTTP status %d", d.ID, resp.StatusCode)); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	return buf.Bytes(), err
}
