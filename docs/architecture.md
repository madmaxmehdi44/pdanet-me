# Architecture

```text
Android VpnService TUN
        |
        | IPv4 packet
        v
PDOP length-prefixed frame
        |
        | adb reverse :10209
        v
Go desktop host
        |
        +-- current milestone: packet responder/test endpoint
        |
        +-- next: TCP/UDP forwarder
```

The key architectural decision is to transport complete IP packets rather than reconstruct TCP state. This allows TCP, UDP, ICMP and later IPv6 to share one packet transport.
