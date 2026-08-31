package com.example.pdanetopen;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.content.Intent;
import android.net.VpnService;
import android.os.Build;
import android.os.IBinder;
import android.os.ParcelFileDescriptor;

import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

public final class OpenVpnService extends VpnService {
    private static final int MTU = 1500;
    private static final int PORT = 10209;
    private static final int MAGIC = 0x50444F50;
    private static final int VERSION = 1;
    private static final int FRAME_IP = 1;
    private static final int MAX_FRAME = 64 * 1024;

    private volatile boolean running;
    private ParcelFileDescriptor tun;
    private Socket socket;
    private Thread txThread;
    private Thread rxThread;

    @Override public IBinder onBind(Intent intent) { return super.onBind(intent); }

    @Override public int onStartCommand(Intent intent, int flags, int startId) {
        if (running) return START_STICKY;
        startForegroundNotification();
        try {
            tun = new Builder()
                    .setSession("PdaNet Open")
                    .setMtu(MTU)
                    .addAddress("10.77.0.2", 24)
                    .addRoute("0.0.0.0", 0)
                    .establish();
            if (tun == null) throw new IOException("VPN establish returned null");

            socket = new Socket();
            if (!protect(socket)) throw new IOException("protect(socket) failed");
            socket.connect(new InetSocketAddress("127.0.0.1", PORT), 5000);

            running = true;
            txThread = new Thread(this::txLoop, "pdanet-open-tx");
            rxThread = new Thread(this::rxLoop, "pdanet-open-rx");
            txThread.start();
            rxThread.start();
        } catch (Exception e) {
            stopSelf();
        }
        return START_STICKY;
    }

    private void txLoop() {
        try (FileInputStream in = new FileInputStream(tun.getFileDescriptor());
             DataOutputStream out = new DataOutputStream(new BufferedOutputStream(socket.getOutputStream()))) {
            byte[] packet = new byte[MTU + 64];
            while (running) {
                int n = in.read(packet);
                if (n <= 0) continue;
                writeFrame(out, FRAME_IP, packet, n);
            }
        } catch (Exception e) {
            shutdown();
        }
    }

    private void rxLoop() {
        try (DataInputStream in = new DataInputStream(new BufferedInputStream(socket.getInputStream()));
             FileOutputStream out = new FileOutputStream(tun.getFileDescriptor())) {
            byte[] header = new byte[12];
            while (running) {
                in.readFully(header);
                ByteBuffer h = ByteBuffer.wrap(header).order(ByteOrder.LITTLE_ENDIAN);
                if (h.getInt() != MAGIC) throw new IOException("bad magic");
                if (h.get() != VERSION) throw new IOException("bad version");
                int type = h.get() & 0xff;
                h.get(); h.get();
                int length = h.getInt();
                if (length < 0 || length > MAX_FRAME) throw new IOException("bad length");
                byte[] payload = new byte[length];
                in.readFully(payload);
                if (type == FRAME_IP) {
                    out.write(payload);
                    out.flush();
                }
            }
        } catch (Exception e) {
            shutdown();
        }
    }

    private static void writeFrame(DataOutputStream out, int type, byte[] payload, int length) throws IOException {
        ByteBuffer h = ByteBuffer.allocate(12).order(ByteOrder.LITTLE_ENDIAN);
        h.putInt(MAGIC).put((byte) VERSION).put((byte) type).put((byte) 0).put((byte) 0).putInt(length);
        out.write(h.array());
        out.write(payload, 0, length);
        out.flush();
    }

    private void startForegroundNotification() {
        final String channelId = "pdanet_open";
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(channelId, getString(R.string.vpn_channel_name), NotificationManager.IMPORTANCE_LOW);
            getSystemService(NotificationManager.class).createNotificationChannel(channel);
        }
        Notification.Builder builder = Build.VERSION.SDK_INT >= Build.VERSION_CODES.O
                ? new Notification.Builder(this, channelId)
                : new Notification.Builder(this);
        startForeground(1, builder.setContentTitle("PdaNet Open").setContentText("VPN tunnel active").setSmallIcon(android.R.drawable.stat_sys_warning).build());
    }

    private void shutdown() {
        running = false;
        try { if (socket != null) socket.close(); } catch (Exception ignored) {}
        try { if (tun != null) tun.close(); } catch (Exception ignored) {}
        socket = null;
        tun = null;
    }

    @Override public void onDestroy() { shutdown(); super.onDestroy(); }
    @Override public void onRevoke() { shutdown(); super.onRevoke(); }
}
