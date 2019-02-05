"""Nextcloud News API implementation."""

import base64
import logging
import time

import dateutil.parser
import flask
import models
from utils import json_response


base_api = flask.Blueprint('nextcloud_base', __name__, url_prefix='/index.php/apps/news/api')
api = flask.Blueprint('nextcloud', __name__, url_prefix='/index.php/apps/news/api/v1-2')


@api.before_request
def authenticate():
    """Authenticate a Nextcloud News user."""
    auth = flask.request.authorization
    if auth:
        if models.authorized_user(auth.username, auth.password):
            return
    return flask.Response('Authentication Required', 401, headers={
            'WWW-Authenticate': 'Basic realm="Login Required"'})


@base_api.route('/')
def index():
    """Index page."""
    return json_response({'apiLevels': ['v1-2']})


@api.route('/user')
def user():
    """Get the status."""
    username = flask.request.authorization.username
    return json_response({
        'userId': username,
        'displayName': username,
        'lastLoginTimestamp': int(time.time()),
        'avatar': None
    })

@api.route('/version')
def version():
    """Get the version."""
    return json_response({'version': '6.0.5'})


@api.route('/status')
def status():
    """Get the status."""
    return json_response({'version': '6.0.5'})


@api.route('/feeds')
def feeds():
    """Get all feeds."""
    feeds = []
    for f in models.Feed.select():
        feeds.append({
            'id': f.id,
            'url': f.url,
            'title': f.name,
            'folderId': f.folder_id,
            'faviconLink': f.faviconLink,
            'link': f.link,
        })
    # TODO: Add newestItemId if there are new items
    return json_response({'feeds': feeds})


@api.route('/folders')
def folders():
    """Get all folders."""
    folders = []
    for f in models.Folder.select():
        folders.append({
            'id': f.id,
            'name': f.name
        })
    return json_response({'folders': folders})


def _item_to_json(i):
    pubDate = dateutil.parser.parse(str(i.pubDate))
    return {
        'id': i.id,
        'guid': i.guid,
        'guidHash': base64.encodebytes(i.guid.encode()).decode().strip(),
        'url': i.url,
        'title': i.title,
        'author': i.author,
        'pubDate': pubDate,
        'body': i.content,
        'feedId': i.feed_id,
        'unread': not i.read,
        'starred': i.starred,
        'lastModified': i.updated,
    }


@api.route('/items')
def items():
    """Get items."""
    batchSize = int(flask.request.args.get('batchSize', -1))
    offset = int(flask.request.args.get('offset', -1))
    getRead = flask.request.args.get('getRead', 'true') == 'true'
    oldestFirst = flask.request.args.get('oldestFirst', 'false') == 'true'
    id_ = int(flask.request.args['id'])
    type_ = int(flask.request.args['type'])
    items = []
    items_query = models.Item.select()
    if type_ == 0:
        items_query = items_query.where(models.Item.feed_id == id_)
    elif type_ == 1:
        feed_ids = [f.id for f in models.Feed.select(models.Feed.id).where(
            models.Feed.folder_id == id_)]
        items_query = items_query.where(models.Item.feed_id.in_(feed_ids))
    elif type_ == 2:
        items_query = items_query.where(models.Item.starred)
    elif type_ == 3:
        pass  # Get all items
    else:
        logging.exception(Exception('Unknown type: {0}'.format(type_)))
        flask.abort(500)
    items = [_item_to_json(i) for i in items_query]
    if offset > 0:
        items_query = items_query.where(models.Item.id <= offset)
    if not getRead:
        items_query = items_query.where(models.Item.read == False)
    if oldestFirst:
        items_query = items_query.order_by(models.Item.pubDate)
    else:
        items_query = items_query.order_by(models.Item.pubDate.desc())
    if batchSize > 0:
        items_query = items_query.limit(batchSize)
    return json_response({'items': items})


@api.route('/items/<int:item_id>/read', methods=['PUT'])
def item_read(item_id):
    """Mark an item as read."""
    i = models.Item.get_by_id(item_id)
    i.read = True
    i.save()
    return json_response({})


@api.route('/items/<int:item_id>/unread', methods=['PUT'])
def item_unread(item_id):
    """Mark an item as unread."""
    i = models.Item.get_by_id(item_id)
    i.read = False
    i.save()
    return json_response({})


@api.route('/feeds/<int:feed_id>/read', methods=['PUT'])
def feed_read(feed_id):
    """Mark items of a feed as read."""
    newestItemId = int(flask.request.args['newestItemId'])
    if feed_id == 0:
        models.Item.update(read=True).where(models.Item.starred).execute()
    else:
        models.Item.update(read=True).where(
            (models.Item.feed_id == feed_id) & (models.Item.id <= newestItemId)).execute()
    return json_response({})


@api.route('/items/updated')
def items_updated():
    """Get updated items."""
    lastModified = int(flask.request.args['lastModified'])
    type_ = int(flask.request.args['type'])
    id_ = int(flask.request.args['id'])
    items_query = None
    if type_ == 0:
        items_query = models.Item.select().where(
            (models.Item.feed_id == id_) & (models.Item.updated >= lastModified)
        )
    elif type_ == 1:
        feed_ids = [f.id for f in models.Feed.select(models.Feed.id).where(
            (models.Feed.folder_id == 1) & (models.Item.updated >= lastModified)
        )]
        items_query = models.Item.select().where(
            (models.Item.feed_id.in_(feed_ids)) & (models.Item.updated >= lastModified)
        )
    elif type_ == 2:
        items_query = models.Item.select().where(
            (models.Item.starred) & (models.Item.updated >= lastModified)
        )
    elif type_ == 3:
        items_query = models.Item.select().where(models.Item.updated >= lastModified)
    else:
        logging.exception(Exception('Unknown type: {0}'.format(type_)))
        flask.abort(500)
    items = [_item_to_json(i) for i in items_query]
    return json_response({'items': items})


@api.route('/items/read/multiple', methods=['PUT'])
def items_read_multiple():
    """Mark multiple items as read."""
    items = flask.request.get_json().get('items', [])
    models.Item.update(read=True).where(models.Item.id.in_(items)).execute()
    return json_response({})


@api.route('/items/unread/multiple', methods=['PUT'])
def items_unread_multiple():
    """Mark multiple items as unread."""
    items = flask.request.get_json().get('items', [])
    models.Item.update(read=False).where(models.Item.id.in_(items)).execute()
    return json_response({})


@api.route('/items/star/multiple', methods=['PUT'])
def items_star_multiple():
    """Mark multiple items as starred."""
    items = flask.request.get_json().get('items', [])
    for i in items:
        feed_id = int(i['feedId'])
        guidHash = i['guidHash']
        guid = base64.decodebytes(guidHash.encode()).decode()
        models.Item.update(starred=True).where(
            (models.Item.feed_id == feed_id) & (models.Item.guid == guid)
        ).execute()
    return json_response({})


@api.route('/items/unstar/multiple', methods=['PUT'])
def items_unstar_multiple():
    """Mark multiple items as unstarred."""
    items = flask.request.get_json().get('items', [])
    for i in items:
        feed_id = int(i['feedId'])
        guidHash = i['guidHash']
        guid = base64.decodebytes(guidHash.encode()).decode()
        models.Item.update(starred=False).where(
            (models.Item.feed_id == feed_id) & (models.Item.guid == guid)
        ).execute()
    return json_response({})
