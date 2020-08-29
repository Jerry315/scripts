#!/usr/bin/env bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python3 /var/virtualenvs/cloud_dns
source /var/virtualenvs/cloud_dns/bin/activate
pip install -r requirements.txt