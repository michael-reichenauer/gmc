@echo off&setlocal

echo Building gmc.exe ...
rem "-s -w" omits debug and symbols to reduse size ("-H=windowsgui" would disable console on windows)
go build -ldflags "-s -w" -o gmc.exe main.go

echo.
echo.
pause