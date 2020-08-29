#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/mapcenter
source /var/virtualenvs/mapcenter/bin/activate
pip install -r requirement.txt