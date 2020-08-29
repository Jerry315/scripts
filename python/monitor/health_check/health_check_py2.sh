#!/bin/bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/health_check/bin/activate
exec python /opt/sa_tools/scripts/py/health_check/health_check_py2.py