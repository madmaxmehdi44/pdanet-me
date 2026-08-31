# PdaNet Open MVP v2

This is a clean, buildable development tree for the packet-level VPN MVP.

## Windows host

Requirements: Go 1.20+.

```powershell
cd desktop
go test ./...
go build -o pdanet-host.exe .\cmd\pdanet-host
.\pdanet-host.exe
```

The host listens on TCP `:10209` by default.

## Android

Requirements: Android Studio with an SDK/API 35 platform and Gradle available on PATH.

```powershell
cd android
.\gradlew.bat assembleDebug
adb install -r .\app\build\outputs\apk\debug\app-debug.apk
```

The Android app requests VPN permission and starts a foreground `VpnService`. For development, the tunnel socket connects to `127.0.0.1:10209`, so use:

```powershell
adb reverse tcp:10209 tcp:10209
```

This version is still a transport/packet-path MVP; it does not yet provide Internet forwarding/NAT. The next milestone replaces the host packet responder with real TCP/UDP forwarding and then Wintun/native USB.
