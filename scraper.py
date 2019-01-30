"""RSS API scraper."""

import datetime
import logging

import feedparser

import models


def scrape_all_feeds():
    """Scrapes all available feeds for new entries."""
    for f in models.Feed.select():
        logging.info('Getting feeds for {0} from {1}'.format(f.name, f.url))
        for e in feedparser.parse(f.url)['entries']:
            if not models.Item.get_or_none(guid=e.guid):
                models.Item.create(
                    guid=e.guid,
                    url=e.link,
                    title=e.title,
                    author=e.author,
                    content=e.content[0]['value'],
                    pubDate=datetime.datetime.fromisoformat(e.updated),
                    feed=f,
                    unread=True,
                    starred=False
                )


if __name__ == '__main__':
    scrape_all_feeds()
