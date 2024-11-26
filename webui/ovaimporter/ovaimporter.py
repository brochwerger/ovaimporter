"""Welcome to Reflex! This file outlines the steps to create a basic app."""

import os
import pathlib
import reflex as rx

from rxconfig import config

from .pages import collectOcpDataPage, ovaUploadingPage

os.environ['REFLEX_UPLOADED_FILES_DIR'] = os.environ.get('REFLEX_UPLOADED_FILES_DIR', '/tmp/ovas')
os.environ['APPROOT'] = str(pathlib.Path().absolute())

app = rx.App()
