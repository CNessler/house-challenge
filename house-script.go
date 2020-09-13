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
	err := getHouses(TotalPages)
	if err != nil {
		log.Printf("Error: %w")
	}
	log.Println("Process Complete!")
}

// getHouses gets all houses by page
func getHouses(pages int) error {
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		houseList, err := tryGetPage(i)
		if err != nil {
			return fmt.Errorf("getHouses: %w", err)
		}
		for _, house := range houseList.Houses {
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
	}

	wg.Wait()
	return nil
}

// tryGetPage gets a single page of houses until successful
func tryGetPage(page int) (*Houses, error) {
	houses, err := getPage(page)
	if err != nil {
		return nil, fmt.Errorf("tryGetPage: %w", err)
	}
	if houses.Ok {
		return houses, nil
	}
	return tryGetPage(page)
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