#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/camera_stats/bin/activate
# [global]
SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
SCRIPT="${SCRIPT_PATH}/main.py"

help=0
period=
store=0

while getopts "hsp:" Option
do case $Option in
    p) period=$OPTARG;;
    s) store=1;;
    h) help=1;;
    esac
done
shift $(($OPTIND - 1))

argv=""

if [ -n $period ]; then
    argv="-p $period"
fi

if [ $store == 1 ]; then
    argv="-s"
fi

if [ help == 1 ]; then
    argv="-h"
fi

exec python $SCRIPT $argv