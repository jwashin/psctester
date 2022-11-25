#VERSION = 2.0
VERSION = 2.02

SNAME = "logwriter.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

import json
import time
import datetime

DATA_VERSION = '1.0'

def now():
    return datetime.datetime.utcnow().isoformat() + 'Z'


class LogWriter(object):

    def __init__(self, filename):
        self.filename = filename
        self.clear()
                
    def clear(self):
        self.overwrite_last = False
        self.data = {'version': DATA_VERSION, 'start':now(), 'messages':[], 'status':'new test'}
        self._write()

    def setstatus(self, status):
        self.data['status'] = status
        self._write()

    def setuser(self, user):
        self.data['user'] = user

    def setfilename(self, name):
        self.data['filename'] = name
        self._write()

    def message(self, s):
        self.data['status'] = 'tests in progress'
        
        if self.overwrite_last:
            self.data['messages'][-1] = s
            self.overwrite_last = False
        else:
            self.data['messages'].append(s)
            
        if s == 'done':
            self.end();
        self._write()

    def changelast(self, s):
        self.data['messages'][-1] = s
        self._write()

    def _write(self):
        with open(self.filename, 'w') as control:
            json.dump(self.data, control, indent=2)

    def setserial(self,s):
        self.data['serial'] = s
        self._write()

    def end(self):
        self.data['end'] = now()
        self.setstatus('done')
         
    def setdata(self, field, value):
        self.data[field] = value
        self._write()

    def getdata(self, field):
        return self.data[field]
    
    def setaddress(self, s):
            self.data['address'] = s
            self._write()

    def volatilemessage(self, msg):
        self.message(msg)
        self.overwrite_last = True
        

thefile = "../control.json"
logwriter = LogWriter(thefile)


if __name__ == '__main__':
    
    #lw = LogWriter('status.json')
    
    lw = logwriter
    
    lw.clear()

    lw.setserial('abc1234')

    lw.message('starting up')

    time.sleep(10)

    lw.message('second stage')

    time.sleep(5)

    lw.message('third stage')
    time.sleep(5)


    lw.setfilename('test.txt')


    lw.message('done')
