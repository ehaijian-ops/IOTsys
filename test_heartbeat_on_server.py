"""Test heartbeat ACK end-to-end by sending a valid WSD heartbeat frame to server."""
import socket
import time

HOST = '127.0.0.1'
PORT = 7000

# Build a valid WSD heartbeat frame (0xA4)
# SOP(1) | LEN(1) | CMD(1) | SESSION_ID(6) | DATA(N) | SUM(1)
sop = 0xEE
session_id = bytes([0x01, 0x02, 0x03, 0x04, 0x05, 0x06])
cmd = 0xA4  # heartbeat

# DATA: PortCount(1) + PortStates(PortCount) + FaultFlag(1) + Voltage(2, LE) + Temp(1)
data = bytes([
    0x01,  # PortCount = 1
    0x00,  # Port 1 state = idle
    0x00,  # FaultFlag = 0
    0x9D, 0x08,  # Voltage = 0x089D LE = 2205 → 220.5V
    0x4B,  # Temperature = 75 → 35°C
])

# LEN = CMD(1) + SID(6) + DATA(N) + SUM(1)
frame_len = 1 + 6 + len(data) + 1  # = 1+6+5+1 = 13
payload = bytes([frame_len, cmd]) + session_id + data
# Calculate XOR
xor = 0
for b in payload:
    xor ^= b
full_frame = bytes([sop]) + payload + bytes([xor])

print(f"Frame ({len(full_frame)} bytes): {full_frame.hex()}")
print(f"  SOP:    0x{full_frame[0]:02X}")
print(f"  LEN:    0x{full_frame[1]:02X} ({full_frame[1]})")
print(f"  CMD:    0x{full_frame[2]:02X}")
print(f"  SID:    {full_frame[3:9].hex()}")
print(f"  DATA:   {full_frame[9:9+len(data)].hex()} ({len(data)} bytes)")
print(f"  SUM:    0x{full_frame[-1]:02X}")

# Connect and send
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.settimeout(5.0)
try:
    sock.connect((HOST, PORT))
    print(f"\nConnected to {HOST}:{PORT}")
    sock.sendall(full_frame)
    print("Heartbeat frame sent, waiting for reply...")
    
    # Read response
    response = sock.recv(1024)
    if response:
        print(f"\n*** RESPONSE RECEIVED ({len(response)} bytes): {response.hex()} ***")
        if len(response) >= 3:
            resp_cmd = response[2]
            if resp_cmd == 0xA5:
                print("  → Heartbeat ACK (0xA5) - SUCCESS!")
            elif resp_cmd == 0xA1:
                print("  → Login ACK (0xA1) - server treated this as login?")
            else:
                print(f"  → Unknown CMD: 0x{resp_cmd:02X}")
    else:
        print("\n*** NO RESPONSE (empty read) ***")
except socket.timeout:
    print("\n*** TIMEOUT: No response received within 5 seconds ***")
except Exception as e:
    print(f"\n*** ERROR: {e} ***")
finally:
    sock.close()
    print("Connection closed.")
