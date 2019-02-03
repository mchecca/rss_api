"""RSS API scraper."""

import datetime
import logging

import feedparser

import models


def scrape_all_feeds():
    """Scrapes all available feeds for new entries."""
    for f in models.Feed.select():
        logging.info('Getting feeds for {0} from {1}'.format(f.name, f.url))
        try:
            fp = feedparser.parse(f.url)
            if 'title' in fp.feed:
                f.title = fp.feed.title
            if 'link' in fp.feed:
                f.link = fp.feed.link
            if 'icon' in fp.feed:
                f.faviconLink = fp.feed.icon.rstrip('/')
            if f.dirty_fields:
                f.save(only=f.dirty_fields)
        except Exception as e:
            logging.exception('Error updating feed')
        for e in fp['entries']:
            try:
                if not models.Item.get_or_none(guid=e.guid):
                    models.Item.create(
                        guid=e.guid,
                        url=e.link,
                        title=e.title,
                        author=e.get('author', 'N/A'),
                        content=e.content[0]['value'],
                        pubDate=datetime.datetime.fromisoformat(e.updated),
                        feed=f,
                        unread=True,
                        starred=False
                    )
            except Exception as ex:
                logging.exception('Error processing feed entry')


if __name__ == '__main__':
    scrape_all_feeds()
