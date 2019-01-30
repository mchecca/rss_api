"""RSS API utility functions."""

import datetime
import json

import flask


def _dthandler(obj):
    """JSON default handler for datetime objects."""
    if isinstance(obj, (datetime.date, datetime.datetime)):
        return int(obj.strftime('%s'))


def json_response(json_obj, status=200, headers=None):
    """Return a JSON response with the specified object and status code."""
    return flask.Response(response=json.dumps(json_obj, default=_dthandler),
                          status=status, mimetype='application/json', headers=headers)
