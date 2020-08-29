# -*- coding: utf-8 -*-
import yaml, os, re
import pandas as pd
from pymongo import MongoClient
from schematics.models import Model



BASE_DIR = os.path.dirname(os.path.abspath(__file__))
config_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(config_file, 'rb'))

session = MongoClient(config['config']['mongodb']['url'])
db = session[config['config']['mongodb']['db']]


class Mawar(Model):
    __database__ = db
    __collection__ = 'mawar_camera'

    def __init__(self, *args, **kwargs):
        super(Mawar, self).__init__(*args, strict=False, **kwargs)
        self._collection = self.__database__.get_collection(self.__collection__)

    @classmethod
    def get_collection(cls):
        '''获取collection对象'''
        return cls.__database__.get_collection(cls.__collection__)

    @classmethod
    def find(cls, filter, fields=None, **kwargs):
        if fields:
            fields = dict([(f, True) for f in fields])
        collection = cls.get_collection()
        records = collection.find(filter, fields, **kwargs)
        return records

    @classmethod
    def update_many(cls, filter, updates):
        '''查找记录，并更新'''
        updates = {'$set': updates}
        collection = cls.get_collection()
        collection.update_many(filter,updates)

    def find_one_and_delete(self, filter):
        '''查找一条记录并删除'''
        self._collection.delete_one(filter)

    @classmethod
    def delete_many(cls, filter):
        collection = cls.get_collection()
        return collection.delete_many(filter)


def convert_to_csv(data):
    df = pd.DataFrame([doc for doc in data])
    df.to_csv('mawar_camera.csv',encoding='utf-8')
    # df.to_json()


def convert_to_json(data):
    # df = pd.DataFrame([doc for doc in data])
    # df.to_json("mawar_camera.json")
    with open("mawar_camera.txt",'wb') as f:
        for doc in data:
            f.write(doc)





if __name__ == '__main__':
    mawar = Mawar()
    filter = {}
    fields = config['config']['fields']
    data = mawar.find(filter,fields)
    convert_to_json(data)


