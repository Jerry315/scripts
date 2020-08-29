#!/bin/bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/ops_crontab/bin/activate

# [global]
SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
SCRIPT_NAME="${SCRIPT_PATH}/mawar.py"
exec python $SCRIPT_NAME
