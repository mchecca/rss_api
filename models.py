"""RSS API Models."""

import datetime

import peewee

import settings


db = peewee.SqliteDatabase(settings.DATABASE_FILE)


def authorized_user(username, password):
    """Check if the specified username and password are authorized."""
    return username in settings.USERS and settings.USERS[username] == password


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

    def auth(self):
        """Get authentication info if applicable."""
        auth = None
        feed = [f for f in settings.FEEDS if f['name'] == self.name][0]
        if 'username' in feed and 'password' in feed:
            auth = (feed['username'], feed['password'])
        return auth


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
for fid, f in enumerate(settings.FOLDERS):
    Folder.get_or_create(name=f['name'])
for fid, f in enumerate(settings.FEEDS):
    folder = Folder.get(name=f['folder'])
    Feed.get_or_create(name=f['name'], url=f['url'], folder=folder)
