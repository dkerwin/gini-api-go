package giniapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Document struct
type Document struct {
	Client *APIClient
	Links  struct {
		Document    string `json:"document"`
		Extractions string `json:"extractions"`
		Layout      string `json:"layout"`
		Processed   string `json:"processed"`
	} `json:"_links"`
	CreationDate         int    `json:"creationDate"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Origin               string `json:"origin"`
	PageCount            int    `json:"pageCount"`
	Pages                []Page `json:"pages"`
	Progress             string `json:"progress"`
	SourceClassification string `json:"sourceClassification"`
}

func (d *Document) String() string {
	return fmt.Sprintf(d.ID)
}

// poll state and return true when done
func (d *Document) Poll(interval time.Duration) bool {
	respChannel := make(chan bool, 1)

	go func() {
		respChannel <- d.WaitForCompletion()
	}()

	select {
	case resp := <-respChannel:
		fmt.Println("Processing state:", resp)
		if resp == true {
			return resp
		}
	case <-time.After(time.Second * interval):
		fmt.Println("Timed out waiting for response!")
		return false
	}

	return true
}

// Update document struct (self)
func (d *Document) Update() Document {
	newDoc := d.Client.Get(d.Links.Document)
	return newDoc
}

func (d *Document) WaitForCompletion() bool {
	for {
		doc := d.Client.Get(d.Links.Document)
		fmt.Println(doc.Progress)
		if doc.Progress == "COMPLETED" || doc.Progress == "ERROR" {
			fmt.Println("Document state ==", doc.Progress)
			return true
		}
	}
	return false
}

// Delete method
func (d *Document) Delete() error {
	resp, err := d.Client.MakeAPIRequest("DELETE", d.Links.Document, nil, "", nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		log.Fatal(resp.Status)
		return err
	}

	return err
}

// Report bug
func (d *Document) ErrorReport(summary string, description string) error {
	resp, err := d.Client.MakeAPIRequest("POST",
		fmt.Sprintf("%s/errorreport?summary=%s&description=%s",
			d.Links.Document,
			summary,
			description,
		), nil, "", nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
		return err
	}

	return err
}

// Layout
func (d *Document) GetLayout() (*Layout, error) {
	var layout Layout

	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Layout, nil, "", nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(contents, &layout)
	if err != nil {
		log.Fatal(err)
	}

	return &layout, err
}

// Extractions
func (d *Document) GetExtractions() (*Extractions, error) {
	var extractions Extractions

	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Extractions, nil, "", nil)
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
		return nil, err
	}
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(contents, &extractions)
	if err != nil {
		log.Fatal(err)
	}

	return &extractions, err
}

// Processed Document
func (d *Document) GetProcessed(filename string) error {
	headers := map[string]string{
		"Accept": "application/octet-stream"
	}
	resp, err := d.Client.MakeAPIRequest("GET", d.Links.Processed, nil, "", headers)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
		return err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("/tmp/processed.pdf", contents, 0644)
	fmt.Println(err)

	return err
}
