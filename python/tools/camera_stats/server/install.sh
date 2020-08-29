#!/usr/bin/env bash
echo "install virtual env"
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
pip install virtualenv
virtualenv -p python /var/virtualenvs/camera_stats
source /var/virtualenvs/camera_stats/bin/activate
pip install -r requirement.txt