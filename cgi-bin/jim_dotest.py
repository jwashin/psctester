import time
import logwriter
import json
import sys
import os
import random
import cgi

def dotest():

    lw = logwriter.logwriter
    z = sys.stdin.read()
    try:
        indata = json.loads(z)
    except:
        indata = {}

    print("indata was {} ({})".format(indata, z))

    address = indata['siteid']

    lw.message('starting up')

    lw.setserial('aabbb333')

    lw.message('starting up 2')

    lw.setuser('pwd')

    lw.message('starting up 3')

##    lw.setaddress(address)

    lw.message('Address = {}'.format(address))

    #time.sleep(600)

    time.sleep(random.randint(1,10))
    for k in range(30):
        lw.message('first stage')

    time.sleep(random.randint(1,10))

    for k in range(30):
        lw.message('second stage')

    time.sleep(random.randint(1,10))

    lw.message('third stage')

    time.sleep(random.randint(1,10))

    filename = 'test.csv'

    lw.setfilename(filename)

    data = ['channel1, 0.345, 0.034', 'channel2, 0.043, 0.44']
    with open('../'+filename,'w') as outfile:
        for k in data:
            outfile.write(k + '\n')

    lw.message('done')
    return True

s = dotest()

print("Content-type: application/json\n")
print(json.dumps({'response':s}))
sys.stdout.flush()
