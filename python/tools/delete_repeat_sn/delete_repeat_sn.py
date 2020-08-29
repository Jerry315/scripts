# -*- coding: utf-8 -*-
import yaml, os, re
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
        # collection.update(filter, updates, **kwargs)

    def find_one_and_delete(self, filter):
        '''查找一条记录并删除'''
        self._collection.delete_one(filter)

    @classmethod
    def delete_many(cls, filter):
        collection = cls.get_collection()
        return collection.delete_many(filter)


if __name__ == '__main__':
    mawar = Mawar()
    # 将重复的cid的sn加上ops_前缀
    # repeat_data = config['config']['repeat']
    # for item in repeat_data:
    #     if item.get('cid', None):
    #         filter = {'sn': item['sn'], '_id': {'$ne': item['cid']}}
    #     else:
    #         filter = {'sn': item['sn']}
    #     update = {'sn': 'ops_'+item['sn']}
    #     mawar.update_many(filter,update)
    # 删除sn前缀为ops_的所有条目
    # filter = {"sn": {"$regex": re.compile('ops')}}
    # docs = mawar.find(filter)
    # for doc in docs:
    #     print doc
    # mawar.delete_many(filter)

