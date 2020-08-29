#!/usr/bin/env bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/elastic_tools
source /var/virtualenvs/elastic_tools/bin/activate
pip install -r requirements.txt