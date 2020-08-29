#!/usr/bin/env bash

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_codispatcher.py"
if [ $# == 0 ]; then
    argv="-q"
fi

if [ $# -ge 1 ]; then
   argv="-p $*"
fi
exec python $script $argv

