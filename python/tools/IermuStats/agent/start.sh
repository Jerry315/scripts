#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/IermuStats/bin/activate
# [global]
SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
SCRIPT="${SCRIPT_PATH}/iermu.py"
exec python $SCRIPT