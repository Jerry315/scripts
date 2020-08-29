#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/check_ssl
source /var/virtualenvs/check_ssl/bin/activate
pip install -r requirement.txt