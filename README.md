# House Challenge
For this code challenge I wrote a single script that will save house images to a local 'photos' directory

### To run locally
1. git clone `git@github.com:CNessler/house-challenge.git`
2. cd to directory and run `./house-script`

### Notes (with more time)
- Given more time I would have added more retry logic utilizing a retry channel and also use an error channel instead of bubbling up the error to be logged. I also would have liked to write some tests for specific methods and get comparisons on download time with and without concurrency. 
