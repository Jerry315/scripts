#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/check_cid_timeout
source /var/virtualenvs/check_cid_timeout/bin/activate
pip install -r requirement.txt