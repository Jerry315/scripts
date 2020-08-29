#!/usr/bin/env bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/IermuStats
source /var/virtualenvs/IermuStats/bin/activate
pip install -r requirement.txt