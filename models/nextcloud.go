package models

import (
	"fmt"
	"time"
)

// JSONTime is a time.Time value formatted for JSON
type JSONTime time.Time

// MarshalJSON converts the JSONTime time.Time value to a Unix timestamp
func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

// NCFeed represents a RSS feed
type NCFeed struct {
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	Title       *string `json:"title"`
	FolderID    int     `json:"folderId"`
	FavIconLink *string `json:"faviconLink"`
	Link        *string `json:"link"`
}

// NCFolder represents a named collection of feeds
type NCFolder struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NCItem represents a RSS feed Item
type NCItem struct {
	ID       int      `json:"id"`
	GUID     string   `json:"guid"`
	GUIDHash string   `json:"guidHash"`
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Author   string   `json:"author"`
	Content  string   `json:"body"`
	PubDate  JSONTime `json:"pubDate"`
	FeedID   int      `json:"feedId"`
	Unead    bool     `json:"unread"`
	Starred  bool     `json:"starred"`
	Updated  JSONTime `json:"lastModified"`
}

func feedsToNc(feeds []Feed) []NCFeed {
	var ncFeeds []NCFeed
	for _, f := range feeds {
		ncFeeds = append(ncFeeds, NCFeed{
			Name:        f.Name,
			URL:         f.URL,
			Title:       f.Title,
			FolderID:    f.FolderID,
			FavIconLink: f.FavIconLink,
			Link:        f.Link,
		})
	}
	return ncFeeds
}

func foldersToNc(folders []Folder) []NCFolder {
	var ncFolders []NCFolder
	for _, f := range folders {
		ncFolders = append(ncFolders, NCFolder{
			ID:   f.ID,
			Name: f.Name,
		})
	}
	return ncFolders
}

func itemsToNc(items []Item) []NCItem {
	var ncItems []NCItem
	for _, i := range items {
		ncItems = append(ncItems, NCItem{
			ID:       i.ID,
			GUID:     i.GUID,
			GUIDHash: i.GUIDHash,
			Title:    i.Title,
			Author:   i.Author,
			Content:  i.Content,
			PubDate:  JSONTime(i.PubDate),
			FeedID:   i.FeedID,
			Unead:    !i.Read,
			Starred:  i.Starred,
			Updated:  JSONTime(i.Updated),
		})
	}
	return ncItems
}

// NCGetAllFeeds gets all feeds for use with NextCloud
func NCGetAllFeeds() []NCFeed {
	return feedsToNc(getAllFeeds())
}

// NCGetAllFolders gets all folders for use with NextCloud
func NCGetAllFolders() []NCFolder {
	return foldersToNc(getAllFolders())
}

// NCQueryItems queries RSS items for use with NextCloud
func NCQueryItems() []NCItem {
	return itemsToNc(queryItems())
}
