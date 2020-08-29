#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_lvs.py"
#if [[ $# < 1 ]]; then
#    exec python ${script} -q
#    exit
#fi
#argv=
#
#if [ $# -eq 2 ]; then
#    argv="-v $1 -p $2"
#fi
#if [ $# -eq 3 ]; then
#    argv="-r $1 -p $2 -k $3"
#fi
#
#if [ $1 == '-h' ]; then
#    argv='-h'
#fi
exec python $script


