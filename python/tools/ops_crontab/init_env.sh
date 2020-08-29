#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/ops_crontab
source /var/virtualenvs/ops_crontab/bin/activate
#/var/virtualenvs/ops_crontab/bin/pip install requests pyyaml
