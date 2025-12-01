@echo off
echo Starting Go microservice...

REM First, ensure no old service is running
call stop_service.bat

REM Compile the latest code
echo Building server.exe...
go build -o server.exe cmd/server/main.go
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)
echo Build successful.

REM Start the server in the background, redirecting output to log files
echo Starting server.exe in background...
start "" server.exe > server.log 2> server_err.log
echo Service started. Check server.log and server_err.log for output.
