package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	DirectoryPermission = 0755
	GetPhotoEndpoint = "http://app-homevision-staging.herokuapp.com/api_project/houses?page="
)

type Houses struct {
	Houses []House `json:"houses"`
	Ok     bool     `json:"ok"`
}
type House struct {
	ID        int    `json:"id"`
	Address   string `json:"address"`
	Homeowner string `json:"homeowner"`
	Price     int    `json:"price"`
	PhotoURL  string `json:"photoURL"`
}

type RetryGetPage struct {
	Error error
	Page int
	RetryCount int
}

func main() {
	log.Println("Starting save houses")
	houseCh := getHouses(10)
	downloadHouse(houseCh)
	log.Println("Process complete!")
	os.Exit(0)
}

// getHouses gets all houses by page
func getHouses(pages int) chan House {
	retryCh := make(chan RetryGetPage)
	houseCh := make(chan House)
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		i := i
		go func() {
			tryGetPage(i, retryCh, houseCh, &wg)
		}()
	}

	go func() {
		for retry := range retryCh {
			wg.Add(1)
			if retry.RetryCount == 5 {
				log.Printf("getHouses: maximum retry limit for page %s", strconv.Itoa(retry.Page))
			}
			time.Sleep(2 * time.Second)
			tryGetPage(retry.Page, retryCh, houseCh, &wg)
			retry.RetryCount++
		}
	}()

	go func() {
		wg.Wait()
		close(retryCh)
		close(houseCh)
	}()

	return houseCh
}

// tryGetPage attempts to get a single page of houses. If successful it puts each house
// on the house channel. If the response is an error or house.Ok is false it puts RetryGetPage
// on the retry channel to try the page again
func tryGetPage(page int, retryCh chan RetryGetPage, houseCh chan House, wg *sync.WaitGroup) {
	defer wg.Done()
	houses, err := getPage(page)
	if err != nil {
		retryCh <- RetryGetPage{
			Error: err,
			Page: page,
		}
		return
	}
	if houses.Ok == false {
		retryCh <- RetryGetPage{
			Error: err,
			Page: page,
		}
		return
	}
	for _, h := range houses.Houses {
		houseCh <- h
	}
	return
}

// getPage gets a single page of Houses from the GetPhotoEndpoint
func getPage(page int) (*Houses, error) {
	res, err := http.Get(GetPhotoEndpoint + strconv.Itoa(page))
	if err != nil {
		return nil, err
	}

	var houses Houses
	err = json.NewDecoder(res.Body).Decode(&houses)
	if err != nil {
		return nil, err
	}
	return &houses, nil
}

// downloadHouse downloads a house to a local file.
// It writes as it downloads without storing the entire image into memory
func downloadHouse(houseCh chan House) {
	_ = os.Mkdir("photos", DirectoryPermission)
	for house := range houseCh {
		house := house
		go func() {
			var response http.Response
			for i := 0; i <= 5; i++ {
				resp, err := http.Get(house.PhotoURL)
				if err != nil {
					if i == 5 {
						log.Printf("error downloading photo for house %s with download url %s", house.ID, house.PhotoURL)
					}
					i++
					time.Sleep(2 * time.Second)
					continue
				}
				defer resp.Body.Close()
				response = *resp
			}

			fileName := fmt.Sprintf("id-[%s]-[%s]%s", strconv.Itoa(house.ID), house.Address, filepath.Ext(house.PhotoURL))
			// Create file
			out, err := os.Create(fmt.Sprintf("./photos/%s", fileName))
			if err != nil {
				log.Printf("downloadHouse: create error for image with filename: %s", fileName)
				return
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, response.Body)
			if err != nil {
				log.Println("downloadHouse: error copying image bytes to file")
				return
			}
		}()
	}
	return
}
