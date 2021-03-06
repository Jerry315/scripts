#!/usr/bin/env bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/monitor_mongo
source /var/virtualenvs/monitor_mongo/bin/activate
pip install -r requirements.txt