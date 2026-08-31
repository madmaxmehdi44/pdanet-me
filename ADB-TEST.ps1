$ErrorActionPreference = 'Stop'
adb devices
adb reverse tcp:10209 tcp:10209
adb reverse --list
Write-Host 'ADB reverse is configured.'
