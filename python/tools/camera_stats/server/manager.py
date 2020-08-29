#!/usr/bin/env python
# -*- coding: utf-8 -*-
from flask import Flask
from api.main import bp


def create_app():
    app = Flask(__name__)
    app.register_blueprint(bp)
    return app


app = create_app()

# if __name__ == '__main__':
#     app.run(host="192.168.2.186", port=5000)
