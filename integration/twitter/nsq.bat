ECHO OFF
ECHO [Running NSQ environment]
start "nsqlookupd" /B nsqlookupd.exe
nsqd.exe --lookupd-tcp-address=localhost:4160
PAUSE