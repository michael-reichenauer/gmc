@echo off&setlocal

echo update dependencies ...
go get -u -t -d -v ./...

echo.
echo.
pause