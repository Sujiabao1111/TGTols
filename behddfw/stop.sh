#!/bin/sh
PID=$(cat app.pid)
echo $PID
kill $PID
rm -f app.pid
