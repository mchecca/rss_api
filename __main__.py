#!/usr/bin/env python3
"""RSS API entrypoint."""

import datetime
import logging
import threading
import time

import flask
import nextcloud
import scraper

app = flask.Flask(__name__)
app.register_blueprint(nextcloud.api)
app.register_blueprint(nextcloud.base_api)


_SCRAPE_DELAY_MINUTES = 15


def _run_scraper():
    while True:
        logging.info('Scraping all feeds')
        try:
            scraper.scrape_all_feeds()
        except Exception as ex:
            logging.exception('Unhandled exception while scraping')
        next_scrape = datetime.datetime.now() + datetime.timedelta(minutes=_SCRAPE_DELAY_MINUTES)
        logging.info('Sleeping for {0} minutes, next scrape at {1}'.format(
            _SCRAPE_DELAY_MINUTES,
            next_scrape
        ))
        time.sleep(_SCRAPE_DELAY_MINUTES * 60)


if __name__ == '__main__':
    host = '0.0.0.0'
    port = 5000
    logging.info('Running RSS API at {0}:{1}'.format(host, port))
    t = threading.Thread(target=_run_scraper, daemon=True)
    t.start()
    app.run(host=host, port=port, debug=False, use_evalex=False)
