# rss-api

RSS Feed Aggregator supporting API access, including Nextcloud News API

## Setup

Use `feeds.yaml` as a starting point and fill it in with your desired feeds. Note that `scrape_interval` is the time in minutes between scraping and `database` is a path to the SQLite database file. Here is an example config file for a Reddit feed

```yaml
database: /data/rss.sqlite
scrape_interval: 10

users:
  - username: my-user
    password: my-password

folders:
  - name: Reddit

feeds:
  - name: Linux
    url: https://www.reddit.com/r/linux/.rss
    folder: Reddit
```

This creates a Nextcloud compatible REST RSS feed database with the Linux Reddit feed secured by the username `my-user` and password `my-password`.

## Running the API

First, build the Docker container
```bash
$ docker build -t rss_api .
```

Make a new directory for the data and copy the `feeds.yaml` file
```bash
$ mkdir data && cp feeds.yaml data/
```

Run the Docker container
```bash
$ docker run --rm -it -p 5000:5000 -v $PWD/data:/data -e RSS_CONFIG_FILE=/data/feeds.yaml rss_api
```

Make a test request
```bash
$ http --auth my-user:my-password GET 127.0.0.1:5000/index.php/apps/news/api/v1-2/feeds
HTTP/1.0 200 OK
Content-Length: 200
Content-Type: application/json
Date: Mon, 14 Oct 2019 01:42:00 GMT
Server: Werkzeug/0.16.0 Python/3.7.4

{
    "feeds": [
        {
            "faviconLink": "https://www.redditstatic.com/icon.png",
            "folderId": 1,
            "id": 1,
            "link": "https://www.reddit.com/r/linux/",
            "title": "Linux",
            "url": "https://www.reddit.com/r/linux/.rss"
        }
    ]
}
```