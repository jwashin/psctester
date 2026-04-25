import socket
import struct
import  fcntl
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sockfd = sock.fileno()
SIOCGIFADDR = 0x8915
import json
import sys
#import cgi

def get_ip(iface = 'eth0'):
    ifreq = struct.pack('16sH14s'.encode('utf-8'), iface.encode('utf-8'), socket.AF_INET, '\x00'.encode('utf-8')*14)
    try:
        res = fcntl.ioctl(sockfd, SIOCGIFADDR, ifreq)
    except:
        return None
    ip = struct.unpack('16sH2x4s8x'.encode('utf-8'), res)[2]
    return socket.inet_ntoa(ip)

mip = {}

logfile = open('testlog.txt',"w")
logfile.write("starting\n")
try:
    for i in ['wlan0','wlan1','eth0','eth1']:
        add = get_ip(i)
        if add:
            mip[i] = add
except Exception as err:
    logfile.write("{}".format(err))




logfile.write(json.dumps(mip))

print ("Content-type: text/json\n\n")
print (json.dumps(mip))
sys.stdout.flush()
