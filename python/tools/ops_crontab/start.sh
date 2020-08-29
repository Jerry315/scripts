#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/ops_crontab/bin/activate
# [global]
SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
CODISKTRACKER="${SCRIPT_PATH}/codisktracker.py"
MAWAR="${SCRIPT_PATH}/mawar.py"
RELAY="${SCRIPT_PATH}/relay.py"
if [ $1 == 'codisktracker' ]
    then
        exec python $CODISKTRACKER
elif [ $1 == 'mawar' ]
    then
        exec python $MAWAR
elif [ $1 == 'relay' ]
    then
        exec python $RELAY
fi