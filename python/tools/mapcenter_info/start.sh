#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/mapcenter/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
cycle=
ip=
port=
peerid=
flag=0
help=0
while getopts ":ahc:i:p:P:" Option
do case $Option in
    a) flag=1;;
    c) cycle=$OPTARG;;
    h) help=1;;
    i) ip=$OPTARG;;
    p) peerid=$OPTARG;;
    P) port=$OPTARG;;
    esac
done
shift $(($OPTIND - 1))

script="${SCRIPT_PATH}/mapcenter_parse.py"
argv=''
if [ $flag == 1 ];then
    argv="${argv} -a"
fi
if [ $help == 1 ];then
    exec python $script -h
fi
if [ -n "$cycle" ]; then
    argv="${argv} -c ${cycle}"
fi
if [ -n "$ip" ]; then
    argv="${argv} -i ${ip}"
fi
if [ -n "$peerid" ]; then
    argv="${argv} -p ${peerid}"
fi
if [ -n "$port" ]; then
    argv="${argv} -P ${port}"
fi
exec python $script $argv