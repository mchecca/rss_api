"""RSS API scraper."""

import datetime
import logging
import pprint
import time

import feedparser
import requests

import models


def scrape_all_feeds():
    """Scrapes all available feeds for new entries."""
    for f in models.Feed.select():
        logging.info('Getting feeds for {0} from {1}'.format(f.name, f.url))
        try:
            auth = f.auth()
            fp_in = f.url
            if auth:
                fp_in = requests.get(f.url, auth=auth).text
            fp = feedparser.parse(fp_in)
            if 'title' in fp.feed:
                f.title = fp.feed.title
            if 'link' in fp.feed:
                f.link = fp.feed.link
            if 'icon' in fp.feed:
                f.faviconLink = fp.feed.icon.rstrip('/')
            if f.dirty_fields:
                f.save(only=f.dirty_fields)
        except Exception as e:
            logging.exception('Error updating feed\n{0}'.format(pprint.pformat(f)))
        for e in fp['entries']:
            try:
                if not models.Item.get_or_none(guid=e.guid):
                    content = content = e.content[0]['value'] if 'content' in e else ''
                    models.Item.create(
                        guid=e.guid,
                        url=e.link,
                        title=e.title,
                        author=e.get('author', 'N/A'),
                        content=content,
                        pubDate=datetime.datetime.fromtimestamp(time.mktime(e.updated_parsed)),
                        feed=f,
                        unread=True,
                        starred=False
                    )
            except Exception as ex:
                logging.exception('Error processing feed entry\n{0}'.format(pprint.pformat(e)))
