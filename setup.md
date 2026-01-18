# 💻 WSL2 + Windows 11 Setup Guide

Follow these steps to set up the recovery environment on your Windows 11 laptop using WSL2.

## 1. Network Configuration (Windows 11)
Since you are connecting directly via UTP cable, you must set a static IP on your **Windows Ethernet Adapter**:

1.  Plug in the UTP cable to the NAS and Laptop.
2.  Open **Settings > Network & Internet > Ethernet**.
3.  Click **Edit** next to **IP assignment**.
4.  Change "Automatic (DHCP)" to **Manual**.
5.  Toggle **IPv4** to **On**.
6.  Enter the following:
    *   **IP address**: `192.168.1.101`
    *   **Subnet mask**: `255.255.255.0`
    *   **Gateway**: (Leave empty)
7.  Click **Save**.

## 2. Transfer Files to WSL
1.  Plug in your Pendrive.
2.  Open your **WSL/Ubuntu terminal**.
3.  Copy the files to your internal Linux home directory (better performance than running from the mount):
    ```bash
    mkdir -p ~/hydra_recovery
    cd ~/hydra_recovery
    # Assuming pendrive is D: drive in Windows
    cp /mnt/d/hydra_v1.0.0.zip .
    unzip hydra_v1.0.0.zip
    cd hydra_dist
    ```

## 3. Set Execution Permissions
Files coming from Windows/Pendrive lose their execution bits. Run this inside the `hydra_dist` folder:
```bash
chmod +x ./parallel.sh
chmod +x ./bin/*
```

## 4. Connectivity Test
Ensure WSL can reach the NAS:
```bash
ping 192.168.1.100
```
*If ping fails, check if your Windows Firewall is blocking the connection.*

## 5. Start Recovery
Verify your settings in `.env` (ensure `HYDRA_URL` is correct), then run:
```bash
./parallel.sh
```

## ♻️ Resuming
If you need to restart your laptop:
1.  Close the terminal.
2.  When you return, open WSL and `cd ~/hydra_recovery/hydra_dist`.
3.  Run `./parallel.sh` again.
4.  Type **`y`** when asked to resume the last session.
