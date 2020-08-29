#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/check_ssl/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/main.py"
if [ $# == 0 ]; then
    argv="-q"
fi

if [ $# == 3 ]; then
    argv="-u $1 -i $2 -p $3 -z"

fi

exec python $script $argv