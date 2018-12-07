package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/reujab/wallpaper"
	"github.com/tidwall/gjson"
)

func main() {
	args := os.Args[1:]

	err := validateSubreddits(args)
	if err != nil {
		fmt.Println(err)
	}

	cycleWallPaper(time.Now(), args)

	for t := range time.NewTicker(24 * time.Hour).C {
		cycleWallPaper(t, args)
	}
}

func cycleWallPaper(tick time.Time, args []string) {
	randomidx := random(0, len(args))
	subreddit := args[randomidx]

	response, err := requestSubreddit(subreddit, true)

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	m, ok := gjson.Parse(string(contents)).Value().(map[string]interface{})
	if !ok {
		fmt.Println("Error")
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}

	children, _, _, err := jsonparser.Get(jsonBytes, "data", "children")
	if len(children) < 20 {
		children = children[:len(children)]
	} else {
		children = children[:20]
	}

	for i := 0; i < len(children)-1; i++ {
		imagePath, _, _, err := jsonparser.Get(jsonBytes, "data", "children", "["+strconv.Itoa(i)+"]", "data", "url")
		if err != nil {
			fmt.Println(err)
		}

		if !strings.Contains(string(imagePath), ".jpg") {
			fmt.Println("url is not a .jpg")
		} else {
			wallpaper.SetFromURL(string(imagePath))
			fmt.Println("Success!")
			fmt.Printf("Wallpaper set to: %s\n", string(imagePath))
			break
		}
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func validateSubreddits(args []string) error {
	for i := 0; i < len(args); i++ {
		resp, err := requestSubreddit(args[i], false)
		if err != nil || resp.StatusCode != 200 {
			fmt.Printf("Invalid subreddit: %s", string(args[i]))
			os.Exit(1)
		}
	}
	return nil
}

func requestSubreddit(subreddit string, json bool) (*http.Response, error) {
	client := &http.Client{}

	var req *http.Request
	var err error
	if json {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://www.reddit.com/r/"+subreddit+".json"), nil)
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://www.reddit.com/r/"+subreddit), nil)
	}
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "script:reddit.reader:v1")

	response, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	return response, err
}
