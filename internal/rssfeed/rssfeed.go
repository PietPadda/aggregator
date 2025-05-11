// rssfeed.go
package rssfeed

import (
	// std go libraries
	"context"      // context for request timeout
	"encoding/xml" // xml unmarshalling
	"fmt"          // printing
	"html"         // html unescaping
	"io"           // file reading
	"net/http"     // http protocol
	"strings"      // checking str contains
)

type RSSFeed struct {
	Channel Channel `xml:"channel"` // Feed channel info
}

type Channel struct {
	Title         string    `xml:"title"`                            // Feed title
	Link          string    `xml:"link"`                             // Feed homepage URL
	Description   string    `xml:"description"`                      // Feed description
	Generator     string    `xml:"generator"`                        // Feed generator name
	Language      string    `xml:"language"`                         // Feed lanauge code
	LastBuildDate string    `xml:"lastBuildDate"`                    // Last update date
	Atom          AtomLink  `xml:"http://www.w3.org/2005/Atom link"` // Atom self URL
	Items         []RSSItem `xml:"item"`                             // Feed items (posts)
}

// HREF shows the actual URL of the feed
// Rel shows relationship, e.g. "self" = the feed itself
// Type shows the MIME type of the feed, e.g. "application/rss+xml"
type AtomLink struct {
	Href string `xml:"href,attr"` // Link URL
	Rel  string `xml:"rel,attr"`  // Relation type
	Type string `xml:"type,attr"` // MIME type
}

type RSSItem struct {
	Title       string `xml:"title"`       // Post title
	Link        string `xml:"link"`        // Post URL
	PubDate     string `xml:"pubDate"`     // Post publication date
	GUID        string `xml:"guid"`        // Unique ID
	Description string `xml:"description"` // Post content
}

// our RSS fetchfeed function
func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// handle empty url
	if feedURL == "" {
		return nil, fmt.Errorf("feed URL is empty")
	}

	// HTTP get request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	// req is the HTTP request to the server

	// HTTP request check
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Header expect rss + xml
	req.Header.Set("Accept", "application/rss+xml")
	// this tells the server that we expect an RSS feed in XML format

	// HTTP client
	client := &http.Client{}
	// client is used to send the HTTP request and get the response

	// set the user agent after request created but before sending the request
	req.Header.Set("User-Agent", "Gator/0.1 (+https://github.com/PietPadda/aggregator)")
	// common practice for web scraping, to identify the client making the request
	// and to avoid being blocked by the server

	// Client do request
	res, err := client.Do(req)
	// res is the client response to the HTTP request

	// Do check
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %w", err)
	}

	// Close response body AFTER confirming non-nil response
	defer res.Body.Close()

	// response content type check
	resType := res.Header.Get("Content-Type")

	// response xml content type check
	if !strings.Contains(resType, "xml") {
		return nil, fmt.Errorf("invalid content type: %s", res.Header.Get("Content-Type"))
	}

	// get server status code
	statusCode := res.StatusCode
	// check if the status code is in the 2xx range

	// error check
	if statusCode > 299 {
		return nil, fmt.Errorf("error fetching feed: %s", res.Status)
	}

	// read the raw response body
	data, err := io.ReadAll(res.Body)
	// io.ReadAll is a helper function to read the response body
	// and return the data as a byte slice

	// read check
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// create RSSFeed instance initialised with empty fields
	// this is the struct that will hold the unmarshalled XML data
	var feed RSSFeed

	// unmarshal XML data into go struct
	err = xml.Unmarshal(data, &feed)
	// it takes the byte slice data and converts it into the RSSFeed struct

	// unmarshal check
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling XML: %w", err)
	}

	// RSSFEED VALIDATION
	// 4 fundamental checks: title, link, description, items > 0
	// check if the feed has a title
	if feed.Channel.Title == "" {
		fmt.Println("Warning: feed has no title")
	}

	// check if the feed has a link
	if feed.Channel.Link == "" {
		fmt.Println("Warning: feed has no link")
	}

	// check if the feed has a description
	if feed.Channel.Description == "" {
		fmt.Println("Warning: feed has no description")
	}

	// check if the feed is empty
	if len(feed.Channel.Items) == 0 {
		fmt.Println("Warning: feed has no items")
	}

	// Unescape the HTML entitites
	// this is done to remove any HTML entities in the channel/feed data
	// Channel Unescape
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Generator = html.UnescapeString(feed.Channel.Generator)
	feed.Channel.Language = html.UnescapeString(feed.Channel.Language)
	feed.Channel.LastBuildDate = html.UnescapeString(feed.Channel.LastBuildDate)
	feed.Channel.Atom.Href = html.UnescapeString(feed.Channel.Atom.Href)
	feed.Channel.Atom.Rel = html.UnescapeString(feed.Channel.Atom.Rel)
	feed.Channel.Atom.Type = html.UnescapeString(feed.Channel.Atom.Type)

	// Items Unescape
	for i := range feed.Channel.Items {
		feed.Channel.Items[i].Title = html.UnescapeString(feed.Channel.Items[i].Title)
		feed.Channel.Items[i].Link = html.UnescapeString(feed.Channel.Items[i].Link)
		feed.Channel.Items[i].PubDate = html.UnescapeString(feed.Channel.Items[i].PubDate)
		feed.Channel.Items[i].GUID = html.UnescapeString(feed.Channel.Items[i].GUID)
		feed.Channel.Items[i].Description = html.UnescapeString(feed.Channel.Items[i].Description)
	}

	// return the feed
	return &feed, nil
	// output is a pointer to the RSSFeed struct
}
