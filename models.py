"""RSS API Models."""

import datetime
import os

import peewee
import yaml

_FOLDER_BASE = os.path.dirname(__file__)
_SETTINGS_FILE = os.environ.get('RSS_CONFIG_FILE', 'feeds.yaml')
assert os.path.exists(_SETTINGS_FILE) and os.access(_SETTINGS_FILE, os.R_OK)

_settings_dict = yaml.load(open(_SETTINGS_FILE, 'r'))
_database_file = os.path.abspath(_settings_dict['database'])

db = peewee.SqliteDatabase(_database_file)

USERS = {u['username']: u['password'] for u in _settings_dict.get('users', {})}


def authorized_user(username, password):
    """Check if the specified username and password are authorized."""
    return username in USERS and USERS[username] == password


class BaseModel(peewee.Model):
    """Base Model for RSS API classes."""

    updated = peewee.DateTimeField(default=datetime.datetime.utcnow())

    def save(self, *args, **kwargs):
        """Save model and set updated field."""
        self.updated = datetime.datetime.utcnow()
        super().save(*args, **kwargs)

    class Meta:
        """Metadata for BaseModel."""

        database = db


class Folder(BaseModel):
    """Folder model."""

    name = peewee.CharField(unique=True)


class Feed(BaseModel):
    """Feed model."""

    name = peewee.CharField(unique=True)
    url = peewee.CharField()
    folder = peewee.ForeignKeyField(Folder)
    link = peewee.CharField(null=True)
    title = peewee.CharField(null=True)
    faviconLink = peewee.CharField(null=True)


class Item(BaseModel):
    """Item model."""

    guid = peewee.CharField(unique=True)
    url = peewee.CharField()
    title = peewee.CharField()
    author = peewee.CharField()
    content = peewee.TextField()
    pubDate = peewee.DateTimeField()
    feed = peewee.ForeignKeyField(Feed)
    read = peewee.BooleanField(default=False)
    starred = peewee.BooleanField(default=False)


db.connect()
db.create_tables([Folder, Feed, Item])

# Make sure all folders and feeds from the settings file are added
for fid, f in enumerate(_settings_dict.get('folders', [])):
    Folder.get_or_create(name=f['name'])
for fid, f in enumerate(_settings_dict.get('feeds', [])):
    folder = Folder.get(name=f['folder'])
    Feed.get_or_create(name=f['name'], url=f['url'], folder=folder)
