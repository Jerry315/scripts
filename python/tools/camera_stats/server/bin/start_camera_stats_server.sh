#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/camera_stats/bin/activate
exec honcho -f procfiles/CameraStatsApi start