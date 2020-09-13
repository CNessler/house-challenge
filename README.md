# House Challenge
For this code challenge I wrote a single script that will save house images to a local 'photos' directory

### To run locally
1. git clone `git@github.com:CNessler/house-challenge.git`
2. cd to directory and run `./houses-challenge`

### Notes (with more time)
- I originally had written this with a houses channel but ran into an issue with the channel never closing. Starting over, I was able to write more readable code by using ioutil instead of io.Copy. I originally used this because the response body will stream instead of writing to disk. However, it was convoluted and difficult to read. Using ioutil also removed a number of lines of uneeded code making it more concise. Given more time I would have added more retry logic instead of erroring out and perhaps change the recursive method to get a page of houses. I also would have liked to write some tests for specific methods and get comparisons on download time with and without concurrency. I think there is room for improvement by also making the page loop a go routine that puts houses on a channel for the download method to pull from. The nested for loop is definitely a code smell that could be cleaned up. 
