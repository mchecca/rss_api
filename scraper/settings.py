"""RSS API settings."""

import os

import yaml

_SETTINGS_FILE = os.environ.get('RSS_CONFIG_FILE', 'feeds.yaml')
assert os.path.exists(_SETTINGS_FILE) and os.access(_SETTINGS_FILE, os.R_OK)

_settings_dict = yaml.load(open(_SETTINGS_FILE, 'r'))

# Settings

DATABASE_FILE = os.path.abspath(_settings_dict['database'])
FOLDERS = _settings_dict.get('folders', [])
FEEDS = _settings_dict.get('feeds', [])
