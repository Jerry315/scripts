#!/usr/bin/env python
# -*- coding: utf-8 -*-
from itsdangerous import TimedJSONWebSignatureSerializer as Serializer

salt = "Wv3hU5oZHcwMPLfg"


class Token(object):
    expire = 300

    def __init__(self, name):
        self.name = name

    def generate_auth_token(self):
        s = Serializer(salt, expires_in=self.expire)
        return s.dumps({"name": self.name})

    def verify_auth_token(self, token):
        s = Serializer(salt)
        try:
            data = s.loads(token)
        except Exception:
            return None
        return data["name"] == self.name

