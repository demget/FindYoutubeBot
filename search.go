package main

import (
	"encoding/json"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type SearchResult struct {
	VideoID      string
	PublishedAt  time.Time `json:"publishedAt"`
	ChannelID    string    `json:"channelId"`
	ChannelTitle string    `json:"channelTitle"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`

	Thumbnail struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
}

func searchVideos(query string) ([]SearchResult, error) {
	params := make(url.Values)
	params.Set("part", "snippet")
	params.Set("q", query)
	params.Set("key", os.Getenv("API_KEY"))
	params.Set("maxResults", "25")
	params.Set("regionCode", "RU")
	params.Set("type", "video")

	url := "https://www.googleapis.com/youtube/v3/search?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var aux struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			}
			Snippet struct {
				SearchResult
				Thumbnails map[string]struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(aux.Items))
	for i, item := range aux.Items {
		results[i] = item.Snippet.SearchResult
		results[i].VideoID = item.ID.VideoID
		results[i].Thumbnail = item.Snippet.Thumbnails["medium"]

		results[i].Title = html.UnescapeString(results[i].Title)
		results[i].ChannelTitle = html.UnescapeString(results[i].ChannelTitle)
	}

	return results, nil
}
