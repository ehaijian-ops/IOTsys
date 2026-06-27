"""Start tcpdump on server to capture heartbeat ACK packets."""
import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('123.56.161.254', username='root', password='hndx@N2000', timeout=15)

# Start tcpdump in background to capture WSD traffic on port 7000
cmd = ('nohup tcpdump -i any -n -XX "port 7000 and host 140.240.31.179" '
       '-w /tmp/heartbeat_capture.pcap '
       '-c 50 '
       '> /tmp/tcpdump.log 2>&1 &')
stdin, stdout, stderr = ssh.exec_command(cmd)
print("tcpdump started:", stdout.read().decode().strip())

# Verify it's running
stdin, stdout, stderr = ssh.exec_command('ps aux | grep "tcpdump.*7000" | grep -v grep')
print("Process:", stdout.read().decode().strip())

ssh.close()
print("\nCapture started. Now waiting for device heartbeats on port 7000...")
print("To check captured packets later:")
print("  ssh root@123.56.161.254 'tcpdump -r /tmp/heartbeat_capture.pcap -XX -n'")
print("To check server logs:")
print("  ssh root@123.56.161.254 'docker logs iot-server --tail 30 -f'")
