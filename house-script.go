package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	DirectoryPermission = 0755
	GetPhotoEndpoint    = "http://app-homevision-staging.herokuapp.com/api_project/houses?page="
	TotalPages          = 10
)

type Houses struct {
	Houses []House
	Ok     bool
}

type House struct {
	ID        int
	Address   string
	Homeowner string
	Price     int
	PhotoURL  string
}

func main() {
	log.Println("Starting process")

	houseCh, err := getHouses(TotalPages)
	if err != nil {
		log.Printf("Error: %w")
	}

	processHouse(houseCh)

	log.Println("Process Complete!")
}

// processHouse takes a channel of House and processes each to be saved
func processHouse(houseCh chan House) {
	var wg sync.WaitGroup
	for house := range houseCh {
		wg.Add(1)
		go func(house House) {
			defer wg.Done()
			readCloser, err := downloadHouse(house)
			if err != nil {
				log.Printf("error downloading house with id %s", strconv.Itoa(house.ID))
			}
			err = writeToDisk(house, readCloser)
			if err != nil {
				log.Printf("error writing house with id %s", strconv.Itoa(house.ID))
			}
		}(house)
	}

	wg.Wait()
	return
}

// getHouses gets all houses by page
func getHouses(pages int) (chan House, error) {
	houseCh := make(chan House)
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		i := i
		go func() {
			tryGetPage(i, houseCh, &wg)
		}()
	}

	go func() {
		wg.Wait()
		close(houseCh)
	}()

	return houseCh, nil
}

// tryGetPage gets a single page of houses until successful
func tryGetPage(page int, houseCh chan House, wg *sync.WaitGroup) {
	houses, err := getPage(page)
	if err != nil {
		log.Printf("tryGetPage: %v", err)
	}
	if houses.Ok {
		for _, h := range houses.Houses {
			houseCh <- h
		}
		wg.Done()
		return
	}
	tryGetPage(page, houseCh, wg)
}

// getPage gets a single page of Houses from the GetPhotoEndpoint
func getPage(page int) (*Houses, error) {
	res, err := http.Get(GetPhotoEndpoint + strconv.Itoa(page))
	if err != nil {
		return nil, fmt.Errorf("getPage: http.Get error %v", err)
	}
	var houses Houses
	err = json.NewDecoder(res.Body).Decode(&houses)
	if err != nil {
		return nil, fmt.Errorf("getPage: json decode errog %v", err)
	}
	return &houses, nil
}

// downloadHouse downloads a house and returns its bytes
func downloadHouse(house House) ([]byte, error) {
	resp, err := http.Get(house.PhotoURL)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("downloadHouse: error downloading photo for house %s with download url %s", strconv.Itoa(house.ID), house.PhotoURL)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("downloadHouse: error reading bytes for house %s", strconv.Itoa(house.ID))
	}
	return bytes, nil
}

// writeToDisk writes a house image to the local photos directory
func writeToDisk(house House, bytes []byte) error {
	_ = os.Mkdir("photos", DirectoryPermission)
	fileName := fmt.Sprintf("id-%s-%s%s", strconv.Itoa(house.ID), house.Address, filepath.Ext(house.PhotoURL))
	err := ioutil.WriteFile(fmt.Sprintf("./photos/%s", fileName), bytes, 777)
	if err != nil {
		return fmt.Errorf("writeToDisk: house %v", house)
	}
	return nil
}