#!/bin/bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"

# [global]
SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
source /var/virtualenvs/ops_crontab/bin/activate
SCRIPT_NAME="${SCRIPT_PATH}/codisktracker.py"
exec python $SCRIPT_NAME
