#!/usr/bin/env python
# -*- coding: utf-8 -*-
from flask import Flask
from start import bp


def create_app():
    app = Flask(__name__)
    app.register_blueprint(bp)
    return app


app = create_app()
