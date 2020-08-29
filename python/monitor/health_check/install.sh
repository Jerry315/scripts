#!/bin/bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/health_check
source /var/virtualenvs/health_check/bin/activate
pip install bs4 requests pyyaml html5lib
