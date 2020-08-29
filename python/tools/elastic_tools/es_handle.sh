#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/elastic_tools/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/es_handle.py"
if [[ $# < 1 ]]; then
    exec python ${script} --help
    exit
fi
[ $# -eq 0 ] && python ${script} --help || python ${script} ${1} ${2}

