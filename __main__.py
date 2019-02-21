#!/usr/bin/env python3
"""RSS API entrypoint."""

import datetime
import logging
import threading
import time

import flask
import prometheus_flask_exporter

import models
import nextcloud
import scraper
import settings

app = flask.Flask(__name__)
app.register_blueprint(nextcloud.api)
app.register_blueprint(nextcloud.base_api)

metrics = prometheus_flask_exporter.PrometheusMetrics(app, path='/metrics')

@app.before_request
def authenticate():
    """Authenticate a Nextcloud News user."""
    auth = flask.request.authorization
    if auth:
        if models.authorized_user(auth.username, auth.password):
            return
    return flask.Response('Authentication Required', 401, headers={
            'WWW-Authenticate': 'Basic realm="Login Required"'})


def _run_scraper():
    while True:
        logging.info('Scraping all feeds')
        try:
            scraper.scrape_all_feeds()
        except Exception as ex:
            logging.exception('Unhandled exception while scraping')
        next_scrape = datetime.datetime.now() + datetime.timedelta(minutes=settings.SCRAPE_INTERVAL)
        logging.info('Sleeping for {0} minutes, next scrape at {1}'.format(
            settings.SCRAPE_INTERVAL,
            next_scrape
        ))
        time.sleep(settings.SCRAPE_INTERVAL * 60)


if __name__ == '__main__':
    host = '0.0.0.0'
    port = 5000
    logging.info('Running RSS API at {0}:{1}'.format(host, port))
    t = threading.Thread(target=_run_scraper, daemon=True)
    t.start()
    metrics.start_http_server(9100, '0.0.0.0')
    app.run(host=host, port=port, debug=False, use_evalex=False)
