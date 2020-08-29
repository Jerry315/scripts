#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/mongo_to_csv
source /var/virtualenvs/mongo_to_csv/bin/activate
pip install pandas==0.23.4 pymongo==3.2.1 schematics==1.1.0 PyYAML==3.13
