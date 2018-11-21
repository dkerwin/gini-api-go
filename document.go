package giniapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	Timing               `json:"-"`
	client               *APIClient
	Owner                string `json:"-"`
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
	return fmt.Sprintf("%s", d.ID)
}

// Poll the progress state of a document and return nil when the processing
// has completed (successful or failed). On timeout return error
func (d *Document) Poll(ctx context.Context, pause time.Duration) APIResponse {
	// store upload duration. Will be overwritten otherwise
	uploadDuration := d.Timing.Upload

	start := time.Now()
	defer func() { d.Timing.Processing = time.Since(start) }()

	type CombinedResponse struct {
		doc  *Document
		resp APIResponse
	}

	docProgress := make(chan CombinedResponse, 1)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				doc, getResponse := d.client.Get(ctx, d.Links.Document, d.Owner)
				docResponse := CombinedResponse{doc, getResponse}

				// we need to keep polling until we hit an error or finish processing
				if getResponse.Error != nil || (doc.Progress == "COMPLETED" || doc.Progress == "ERROR") {
					docProgress <- docResponse
					return
				}

				// be a (potentially) good neighbour
				time.Sleep(pause)
			}
		}
	}(ctx)

	select {
	case answer := <-docProgress:
		if answer.doc == nil || answer.resp.Error != nil {
			return answer.resp
		}

		// replace ourself with the polled document
		*d = *answer.doc

		// restore upload duration
		d.Timing.Upload = uploadDuration

		return apiResponse("polling completed", d.ID, answer.resp.HttpResponse, nil)
	case <-ctx.Done():
		return apiResponse("polling aborted", d.ID, nil, ctx.Err())
	}
}

// Update document struct from self-contained document link
func (d *Document) Update(ctx context.Context) APIResponse {
	newDoc, resp := d.client.Get(ctx, d.Links.Document, d.Owner)

	if resp.Error != nil {
		return resp
	}

	*d = *newDoc

	return apiResponse("update completed", d.ID, resp.HttpResponse, nil)
}

// Delete a document
func (d *Document) Delete(ctx context.Context) APIResponse {
	resp, err := d.client.makeAPIRequest(ctx, "DELETE", d.Links.Document, nil, nil, d.Owner)

	if err != nil {
		return apiResponse(ErrHTTPDeleteFailed, d.ID, resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiResponse(ErrDocumentDelete, d.ID, resp, errors.New(ErrDocumentDelete))
	}

	return apiResponse("delete completed", d.ID, resp, nil)
}

// GetLayout returns the JSON representation of a documents layout parsed as
// Layout struct
func (d *Document) GetLayout(ctx context.Context) (*Layout, APIResponse) {
	var layout Layout

	resp, err := d.client.makeAPIRequest(ctx, "GET", d.Links.Layout, nil, nil, "")

	if err != nil {
		return nil, apiResponse(ErrHTTPGetFailed, d.ID, resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiResponse(ErrDocumentLayout, d.ID, resp, errors.New(ErrDocumentLayout))
	}

	if err := json.NewDecoder(resp.Body).Decode(&layout); err != nil {
		return nil, apiResponse("decoding failed", d.ID, resp, err)
	}

	return &layout, apiResponse("layout completed", d.ID, resp, err)
}

// GetExtractions returns a documents extractions in a Extractions struct
func (d *Document) GetExtractions(ctx context.Context, incubator bool) (*Extractions, APIResponse) {
	var extractions Extractions
	var headers map[string]string

	if incubator {
		headers = map[string]string{
			"Accept": "application/vnd.gini.incubator+json",
		}
	}

	resp, err := d.client.makeAPIRequest(ctx, "GET", d.Links.Extractions, nil, headers, d.Owner)

	if err != nil {
		return nil, apiResponse(ErrHTTPGetFailed, d.ID, resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiResponse(ErrDocumentExtractions, d.ID, resp, errors.New(ErrDocumentExtractions))
	}

	if err := json.NewDecoder(resp.Body).Decode(&extractions); err != nil {
		return nil, apiResponse("decoding failed", d.ID, resp, err)
	}

	return &extractions, apiResponse("extractions completed", d.ID, resp, err)
}

// GetProcessed returns a byte array of the processed (rectified, optimized) document
func (d *Document) GetProcessed(ctx context.Context) ([]byte, APIResponse) {
	headers := map[string]string{
		"Accept": "application/octet-stream",
	}

	resp, err := d.client.makeAPIRequest(ctx, "GET", d.Links.Processed, nil, headers, d.Owner)

	if err != nil {
		return nil, apiResponse(ErrHTTPGetFailed, d.ID, resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiResponse(ErrDocumentProcessed, d.ID, resp, errors.New(ErrDocumentProcessed))
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)

	if err != nil {
		return nil, apiResponse(ErrDocumentProcessed, d.ID, resp, err)
	}

	return buf.Bytes(), apiResponse("processed completed", d.ID, resp, err)
}

// SubmitFeedback submits feedback from map
func (d *Document) SubmitFeedback(ctx context.Context, feedback map[string]map[string]interface{}) APIResponse {
	feedbackMap := map[string]map[string]map[string]interface{}{
		"feedback": feedback,
	}

	feedbackBody, err := json.Marshal(feedbackMap)
	if err != nil {
		return apiResponse("encoding failed", d.ID, nil, err)
	}

	resp, err := d.client.makeAPIRequest(ctx, "PUT", d.Links.Extractions, bytes.NewReader(feedbackBody), nil, d.Owner)

	if err != nil {
		return apiResponse(ErrHTTPPutFailed, d.ID, resp, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiResponse(ErrDocumentFeedback, d.ID, resp, errors.New(ErrDocumentFeedback))
	}

	return apiResponse("feedback completed", d.ID, resp, err)
}
