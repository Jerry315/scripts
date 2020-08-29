# -*- coding: utf-8 -*-
from flask import Blueprint, jsonify
from pymongo import MongoClient
from common import config
from common import BaseModel, Logger

bp = Blueprint('camera_stats', __name__, url_prefix="/camera/stats/v1")
camera_fields = config['mongodb']['camera']['fields']
camera_collection = config['mongodb']['camera']['table']
camera_session = MongoClient(config['mongodb']['camera']['url'])
camera_db = camera_session[config['mongodb']['camera']['db']]

device_fields = config['mongodb']['devices']['fields']
device_collection = config['mongodb']['devices']['table']
device_session = MongoClient(config['mongodb']['devices']['url'])
device_db = device_session[config['mongodb']['devices']['db']]

app_fields = config['mongodb']['mawarapp']['fields']
app_collection = config['mongodb']['mawarapp']['table']
app_session = MongoClient(config['mongodb']['mawarapp']['url'])
app_db = app_session[config['mongodb']['mawarapp']['db']]


class AppModel(BaseModel):
    __database__ = app_db
    __collection__ = app_collection


class CameraModel(BaseModel):
    __database__ = camera_db
    __collection__ = camera_collection


class DevicesModel(BaseModel):
    __database__ = device_db
    __collection__ = device_collection


@bp.route("/cid_info")
def camera():
    camera_handle = CameraModel()
    records = camera_handle.find({}, camera_fields)
    record_list = [record for record in records]
    data = {"data": record_list}
    return jsonify(data)


@bp.route("/device_info")
def device():
    device_handle = DevicesModel()
    app_handle = AppModel()
    device_records = device_handle.find({}, device_fields)
    device_record_list = [record for record in device_records]
    record_list = []
    step = 2000
    dl = len(device_record_list)
    d, m = divmod(dl, step)
    if m > 0:
        d = d + 1
    for i in range(d):
        if (i + 1) * step > dl:
            end = dl
        else:
            end = (i + 1) * step
        start = i * step
        records = device_record_list[start:end]
        cids = [record["_id"] for record in records]
        app_records = app_handle.find({"_id": {"$in": cids}}, app_fields)
        for record in records:
            record["group"] = ""
            record["is_bind"] = False
            for ar in app_records:
                if record["_id"] == ar["_id"]:
                    record["group"] = ar["group"]
                    if ar.get("is_bind", False):
                        record["is_bind"] = ar["is_bind"]
                    else:
                        record["is_bind"] = False
                    break
            record_list.append(record)
    data = {"data": record_list}
    return jsonify(data)


if __name__ == '__main__':
    logger = Logger()
    try:
        from flask import Flask

        app = Flask(__name__)
        app.register_blueprint(bp)
        host = config['host']
        port = config['port']
        app.run(host=host, port=port)
    except Exception as e:
        logger.error(str(e))
