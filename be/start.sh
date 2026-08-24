#!/bin/bash
go_exe='/root/path/to/yourbin'
ps -fe|grep $go_exe |grep -v grep >/dev/null
if [ $? -eq 0 ]
then
	echo "already running!"
	exit
else
	echo "appllication start..."
fi
#nohup $go_exe >/dev/null 2>&1 &
nohup $go_exe &
app_pid=0
if [ $? -eq 0 ]
then
	app_pid=$!
else
	echo "error"
	exit
fi
echo $app_pid
echo $app_pid > app.pid
echo "ok"
