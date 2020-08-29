#!/usr/bin/env bash

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_codproducer.py"
if [ $# == 0 ]; then
    argv="-q"
fi

if [ $# -ge 1 ]; then
   argv="-p "
   for key in $@
   do
       argv="${argv}${key},"
    done
fi

exec python $script $argv

