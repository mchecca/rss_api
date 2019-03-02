package models

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"time"

	"database/sql"

	// Blank import needed to use Go SQLite
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type userYaml struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type folderYaml struct {
	Name string `yaml:"name"`
}

type feedYaml struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Folder string `yaml:"folder"`
}

// SettingsYaml is the settings representing by the YAML file
type SettingsYaml struct {
	Database string       `yaml:"database"`
	Users    []userYaml   `yaml:"users"`
	Folders  []folderYaml `yaml:"folders"`
	Feeds    []feedYaml   `yaml:"feeds"`
}

var settings SettingsYaml

func init() {
	configFilePath, exists := os.LookupEnv("RSS_CONFIG_FILE")
	if !exists {
		panic("Environment variable RSS_CONFIG_FILE not set")
	}
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(configFile, &settings)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	// Ensure DB schema
	db := getDB()
	defer db.Close()
	tx, err := db.Begin()
	tx.Exec("CREATE TABLE IF NOT EXISTS \"folder\" (\"id\" INTEGER NOT NULL PRIMARY KEY, \"updated\" DATETIME NOT NULL, \"name\" VARCHAR(255) NOT NULL);")
	tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS \"folder_name\" ON \"folder\" (\"name\");")
	tx.Exec("CREATE TABLE IF NOT EXISTS \"feed\" (\"id\" INTEGER NOT NULL PRIMARY KEY, \"updated\" DATETIME NOT NULL, \"name\" VARCHAR(255) NOT NULL, \"url\" VARCHAR(255) NOT NULL, \"folder_id\" INTEGER NOT NULL, \"link\" VARCHAR(255), \"title\" VARCHAR(255), \"faviconLink\" VARCHAR(255), FOREIGN KEY (\"folder_id\") REFERENCES \"folder\" (\"id\"));")
	tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS \"feed_name\" ON \"feed\" (\"name\");")
	tx.Exec("CREATE INDEX IF NOT EXISTS \"feed_folder_id\" ON \"feed\" (\"folder_id\");")
	tx.Exec("CREATE TABLE IF NOT EXISTS \"item\" (\"id\" INTEGER NOT NULL PRIMARY KEY, \"updated\" DATETIME NOT NULL, \"guid\" VARCHAR(255) NOT NULL, \"url\" VARCHAR(255) NOT NULL, \"title\" VARCHAR(255) NOT NULL, \"author\" VARCHAR(255) NOT NULL, \"content\" TEXT NOT NULL, \"pubDate\" DATETIME NOT NULL, \"feed_id\" INTEGER NOT NULL, \"read\" INTEGER NOT NULL, \"starred\" INTEGER NOT NULL, FOREIGN KEY (\"feed_id\") REFERENCES \"feed\" (\"id\"));")
	tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS \"item_guid\" ON \"item\" (\"guid\");")
	tx.Exec("CREATE INDEX IF NOT EXISTS \"item_feed_id\" ON \"item\" (\"feed_id\");")
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Unable to commit transaction: %v", err)
	}
	// Add folders and feeds
	now := time.Now().Unix()
	var row *sql.Row
	for fid, f := range settings.Folders {
		var count int
		row = db.QueryRow("SELECT COUNT(1) FROM folder WHERE name = ?", f.Name)
		err = row.Scan(&count)
		if err != nil {
			log.Fatalf("Unable to get count: %v\n", err)
		}
		if count == 0 {
			_, err = db.Exec("INSERT INTO folder(id, name, updated) VALUES(?,?,?)", fid, f.Name, now)
			if err != nil {
				log.Fatalf("Unable to insert folder: %v\n", err)
			}
		}
	}
	for fid, f := range settings.Feeds {
		var folderID int
		row = db.QueryRow("SELECT id FROM folder WHERE name = ?", f.Folder)
		err = row.Scan(&folderID)
		if err != nil {
			log.Fatalf("Unable to get folder: %v\n", err)
		}
		var count int
		row = db.QueryRow("SELECT COUNT(1) FROM feed WHERE name = ? AND url = ? AND folder_id = ?", f.Name, f.URL, folderID)
		err = row.Scan(&count)
		if err != nil {
			log.Fatalf("Unable to get count: %v\n", err)
		}
		if count == 0 {
			_, err = db.Exec("INSERT INTO feed(id, name, folder_id, url, updated) VALUES(?,?,?,?,?)", fid, f.Name, folderID, f.URL, now)
			if err != nil {
				log.Fatalf("Unable to insert feed: %v\n", err)
			}
		}
	}
	log.Printf("Using database file: %s\n", settings.Database)
}

// Feed represents a RSS feed
type Feed struct {
	Name        string
	URL         string
	Title       *string
	FolderID    int
	FavIconLink *string
	Link        *string
}

// Folder represents a named collection of feeds
type Folder struct {
	ID   int
	Name string
}

// Item represents a RSS feed Item
type Item struct {
	ID       int
	GUID     string
	GUIDHash string
	URL      string
	Title    string
	Author   string
	Content  string
	PubDate  time.Time
	FeedID   int
	Read     bool
	Starred  bool
	Updated  time.Time
}

// AuthenticateUser checks the given username and password to authenticate the user
func AuthenticateUser(username string, password string) bool {
	for _, user := range settings.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func getDB() *sql.DB {
	db, err := sql.Open("sqlite3", settings.Database)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	return db
}

func runQuery(db *sql.DB, query string) *sql.Rows {
	rows, err := db.Query(query)
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return rows
}

func getAllFeeds() []Feed {
	var feeds []Feed
	db := getDB()
	defer db.Close()
	rows := runQuery(db, "SELECT name, url, title, folder_id, faviconLink, link FROM feed;")
	defer rows.Close()
	for rows.Next() {
		var name string
		var url string
		var title *string
		var folderID int
		var faviconLink *string
		var link *string
		err := rows.Scan(&name, &url, &title, &folderID, &faviconLink, &link)
		if err != nil {
			log.Fatal(err)
		}
		feeds = append(feeds, Feed{
			Name:        name,
			URL:         url,
			Title:       title,
			FolderID:    folderID,
			FavIconLink: faviconLink,
			Link:        link,
		})
	}
	return feeds
}

func getAllFolders() []Folder {
	var folders []Folder
	db := getDB()
	defer db.Close()
	rows := runQuery(db, "SELECT id, name FROM folder;")
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		folders = append(folders, Folder{
			ID:   id,
			Name: name,
		})
	}
	return folders
}

func queryItems() []Item {
	var items []Item
	db := getDB()
	defer db.Close()
	rows := runQuery(db, "SELECT id, guid, url, title, author, content, pubDate, feed_id, read, starred, updated FROM item;")
	defer rows.Close()
	for rows.Next() {
		var id int
		var guid string
		var url string
		var title string
		var author string
		var content string
		var pubDate time.Time
		var feedID int
		var read bool
		var starred bool
		var updated time.Time
		err := rows.Scan(&id, &guid, &url, &title, &author, &content, &pubDate, &feedID, &read, &starred, &updated)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, Item{
			ID:       id,
			GUID:     guid,
			GUIDHash: base64.StdEncoding.EncodeToString([]byte(guid)),
			URL:      url,
			Title:    title,
			Author:   author,
			Content:  content,
			PubDate:  pubDate,
			FeedID:   feedID,
			Read:     read,
			Starred:  starred,
			Updated:  updated,
		})
	}
	return items
}
