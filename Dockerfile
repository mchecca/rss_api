FROM python:3-alpine
ENV PYTHONPATH=/app
COPY requirements.txt /app/rss_api/requirements.txt
RUN pip install -r /app/rss_api/requirements.txt
WORKDIR /app/rss_api
COPY . /app/rss_api
CMD ["python3", "-m", "rss_api"]
