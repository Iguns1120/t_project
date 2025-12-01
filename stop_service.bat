@echo off
echo Stopping Go microservice...

REM Find and kill processes listening on port 8080
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080') do (
    if "%%a" neq "" (
        echo Killing PID: %%a (on port 8080)
        taskkill /F /PID %%a 2>nul
    )
)

REM Kill specific process names if they are still running (e.g., from old go run commands)
taskkill /F /IM server.exe 2>nul
taskkill /F /IM go.exe 2>nul

echo Service stopped.
