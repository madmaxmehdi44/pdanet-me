@echo off
setlocal
where gradle >nul 2>nul
if %ERRORLEVEL%==0 (
  gradle %*
  exit /b %ERRORLEVEL%
)
echo Gradle was not found on PATH.
echo Install Android Studio/Gradle, or add gradle.bat to PATH, then run this command again.
exit /b 1
