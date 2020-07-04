@echo off&setlocal

echo Building docker container ...
docker build -q -t gmcbuilder -f ./build/Dockerfile_gmcbuilder build 
if %ERRORLEVEL% neq 0 goto :ERROR
docker run -it --rm --mount type=bind,source="%CD%",target=/app -w /app gmcbuilder /app/gmc_linux
if %ERRORLEVEL% neq 0 goto :ERROR



pause
EXIT /B %errorlevel%

:ERROR
echo.
echo ERROR has occurred !!
pause