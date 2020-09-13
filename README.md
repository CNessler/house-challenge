# House Challenge
For this code challenge I wrote a single script that will save house images to a local 'photos' directory

## To run locally
1. git clone `git@github.com:CNessler/house-challenge.git`
2. cd to directory and run `go run house-script.go`

### Notes (with more time)
- I would like to have added tests specifically around the retry logic 
- It would have been nice to have some comparisons for download time with and without having concurrency
  - What would be the difference if the pagination calls were also concurrent?
