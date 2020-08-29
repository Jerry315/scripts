# -*- coding: utf-8 -*-
import json
import time
from schematics.models import Model
from pymongo import ReturnDocument, MongoClient
from settings import config

# print config
mongo_config = config['mongodb']
db_session = MongoClient(mongo_config['url'], connect=False,)
db = db_session[mongo_config['db']]


class BaseModel(Model):
    __database__ = ''
    __collection__ = ''

    def __init__(self, **kwargs):
        super(BaseModel, self).__init__(**kwargs)
        self._collection = self.__database__.get_collection(self.__collection__)

    def to_json(self):
        return json.dumps(self.serialize())

    def to_dict(self, fields, func=None):
        dct = {}
        for field in fields:
            value = self.get(field)
            if func:
                value = func(value)
            dct[field] = value
        return dct

    @classmethod
    def next_id(self):
        filter = {'_id': self.__collection__}
        update = {'$inc': {'seq': 1}}
        rv = self.__database__.counters.find_one_and_update(
            filter, update, upsert=True, return_document=ReturnDocument.AFTER
        )
        return rv['seq']

    @classmethod
    def get_collection(cls):
        return cls.__database__.get_collection(cls.__collection__)

    @classmethod
    def distinc(cls, field, filter=None):
        collection = cls.get_collection()
        rv = collection.distinct(field, filter)
        return rv

    @classmethod
    def find_by_id(cls, _id, fields=None):
        filter = {'_id': _id}
        if fields:
            fields = dict([(f, True) for f in fields])
        collection = cls.__database__.get_collection(cls.__collection__)
        record = collection.find_one(filter, fields)
        if not record:
            return None
        return cls(record)

    @classmethod
    def find_one(cls, filter, fields=None, add_empty=False, **kwargs):
        if fields:
            fields = dict([(f, True) for f in fields])
        collection = cls.__database__.get_collection(cls.__collection__)
        record = collection.find_one(filter, fields, **kwargs)
        if not record:
            if add_empty:
                return cls({})
            return None
        return cls(record)

    @classmethod
    def find_last_one(cls, filter, time_filed, fileds=None, time_limit=None):
        if fileds:
            fileds = dict([(f, True) for f in fileds])
        collection = cls.__database__.get_collection(cls.__collection__)
        if time_limit:
            filter.update({'create_time': {'$gte': time_limit}})
        record = collection.find_one(
            filter, fileds, sort=[(time_filed, -1)]
        )
        if not record:
            return {}
        return record

    @classmethod
    def find(cls, filter, fields=None, **kwargs):
        if fields:
            fields = dict([(f, True) for f in fields])
        collection = cls.__database__.get_collection(cls.__collection__)
        records = collection.find(filter, fields, **kwargs)
        return records

    @classmethod
    def find_by_ids(cls, ids, fields=None):
        return cls._find_by_field_data('_id', ids, fields)

    @classmethod
    def find_one_and_update(cls, filter, updates, fields=None, upsert=False,
                            return_doc=ReturnDocument.BEFORE, set_on_insert=None):
        if fields:
            fields = dict([(f, True) for f in fields])
        if set_on_insert:
            updates['$setOnInsert'] = set_on_insert
        updates['update_time'] = time.time()
        updates = {'$set': updates}
        collection = cls.get_collection()
        record = collection.find_one_and_update(
            filter, updates, projection=fields, return_document=return_doc, upsert=upsert
        )
        if not record:
            return None
        return record

    @classmethod
    def _find_by_field_data(cls, field, data, fields=None):
        if fields:
            fields = dict([(f, True) for f in fields])
        filter = {field: {'$in': data}}
        collection = cls.get_collection()
        records = collection.find(filter, fields)
        return records

    @classmethod
    def _find_one_by_field_data(cls, field, data, fields=None):
        if fields:
            fields = dict([(f, True) for f in fields])
        filter = {field: {'$in': data}}
        collection = cls.get_collection()
        record = collection.find_one(filter, fields)
        if not record:
            return None
        return record

    @classmethod
    def insert(cls, doc):
        collection = cls.__database__.get_collection(cls.__collection__)
        collection.insert(doc)

    @classmethod
    def multi_insert(cls, docs):
        collection = cls.__database__.get_collection(cls.__collection__)
        collection.insert_many([dict(doc.items()) for doc in docs])

    @classmethod
    def total(cls, filter={}):
        collection = cls.__database__.get_collection(cls.__collection__)
        return collection.count(filter)


class DeviceModel(BaseModel):
    __database__ = db
    __collection__ = 'camera_stats'
