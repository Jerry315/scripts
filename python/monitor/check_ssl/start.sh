#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/check_ssl/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
expire=
ip=
port=
url=
flag=0
help=0
zabbix=0
query=0
while getopts ":dhqze:i:p:u:" Option
do case $Option in
    d) flag=1;;
    e) expire=$OPTARG;;
    h) help=1;;
    i) ip=$OPTARG;;
    u) url=$OPTARG;;
    p) port=$OPTARG;;
    q) query=1;;
    z) zabbix=1;;
    esac
done
shift $(($OPTIND - 1))

script="${SCRIPT_PATH}/main.py"
argv=''
if [ $flag -eq 1 ];then
    argv="${argv} -d"
fi

if [ $zabbix -eq 1 ];then
    argv="${argv} -z"
fi

if [ $query -eq 1 ];then
    argv="${argv} -q"
fi

if [ $help -eq 1 ];then
    exec python $script -h
fi

if [ -n "$expire" ]; then
    argv="${argv} -c ${expire}"
fi

if [ -n "$ip" ]; then
    argv="${argv} -i ${ip}"
fi

if [ -n "$url" ]; then
    argv="${argv} -u ${url}"
fi

if [ -n "$port" ]; then
    argv="${argv} -p ${port}"
fi
exec python $script $argv