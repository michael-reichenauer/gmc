@echo off&setlocal

echo Building docker container ...
docker build -q -t gmcbuilder -f ./build/Dockerfile_gmcbuilder build 
if %ERRORLEVEL% neq 0 goto :ERROR
echo Building gmc ...
docker run -it --rm --mount type=bind,source="%CD%",target=/app -w /app gmcbuilder go run /app/build/build.go
if %ERRORLEVEL% neq 0 goto :ERROR

rundll32 url.dll,FileProtocolHandler https://github.com/michael-reichenauer/gmc/releases/new
if %ERRORLEVEL% neq 0 goto :ERROR

pause
EXIT /B %errorlevel%

:ERROR
echo.
echo ERROR has occurred !!
pause