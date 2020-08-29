#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/check_cid_timeout/bin/activate


SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/check_cid_timeout.py"
ps_name=$*
python ${script} ${ps_name}