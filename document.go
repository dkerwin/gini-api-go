package giniapi

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

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
	resp, err := d.Client.MakeAPIRequest("DELETE", d.Links.Document, nil)
	if resp.StatusCode != http.StatusNoContent {
		log.Fatal(resp.Status)
		return err
	}
	if err != nil {
		log.Fatal(err)
		return err
	}

	return err
}
