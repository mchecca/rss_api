FROM python:3-alpine
ENV PYTHONPATH=/app RSS_CONFIG_FILE=/data/feeds.yaml
WORKDIR /app/rss_api_scraper
COPY requirements.txt /app/rss_api/requirements.txt
RUN pip install -r /app/rss_api/requirements.txt
COPY . /app/rss_api_scraper
CMD ["python3", "-m", "rss_api_scraper"]
