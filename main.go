package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type Gif struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Images struct {
		Original struct {
			URL string `json:"url"`
		} `json:"original"`
	} `json:"images"`
}

type SearchResult struct {
	Data       []Gif `json:"data"`
	Pagination struct {
		TotalCount int `json:"total_count"`
	} `json:"paginator"`
}

type GiphyService struct {
	BaseUrl string
	APIKey  string
	Dir     string
}

func (g GiphyService) Search(query string, limit int) ([]Gif, error) {
	gifs := []Gif{}

	url := fmt.Sprintf("%s/gifs/search?api_key=%s&q=%s&limit=%d", g.BaseUrl, g.APIKey, query, limit)

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	searchResult := SearchResult{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&searchResult)

	if err != nil {
		return gifs, err
	}

	return searchResult.Data, nil
}

func (g GiphyService) Download(gifs []Gif) {
	fmt.Printf("Downloading gifs to: %s\n", g.Dir)

	var wg sync.WaitGroup
	for _, gif := range gifs {
		wg.Add(1)

		go func(gif Gif) {
			defer wg.Done()

			output, err := os.Create(fmt.Sprintf("%s/%s.gif", g.Dir, gif.Slug))

			if err != nil {
				fmt.Println("Error while creating gif file:", gif.Slug, err)
				return
			}

			defer output.Close()

			url := gif.Images.Original.URL
			fmt.Printf("Dowloading %s...\n", url)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error while downloading...")
				return
			}

			defer resp.Body.Close()

			_, err = io.Copy(output, resp.Body)
			if err != nil {
				fmt.Println("Error while write file")
			}
		}(gif)
	}

	wg.Wait()
	fmt.Println("Download done!\n")
}

func main() {
	apiKey := os.Getenv("API_KEY")

	if apiKey == "" {
		log.Fatal("API_KEY env. variable is required")
	}

	var limit int
	var dir string
	flag.IntVar(&limit, "limit", 5, "maximum number of gifs to download")
	flag.StringVar(&dir, "dir", "./gifs", "gifs download path")
	flag.Parse()

	gs := GiphyService{
		BaseUrl: "http://api.giphy.com/v1",
		APIKey:  apiKey,
		Dir:     dir,
	}

	if len(flag.Args()) == 0 {
		log.Fatal("search query is required")
	}

	query := flag.Args()[0]

	gifs, err := gs.Search(query, limit)
	if err != nil {
		log.Fatal(err)
	}

	gs.Download(gifs)
}
