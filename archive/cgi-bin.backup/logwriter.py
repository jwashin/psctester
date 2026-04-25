#VERSION = 2.0
import os
import datetime
import time
import json
VERSION = 2.02

SNAME = "logwriter.py"

ver_file = open('./script_ver_work', 'a')
ver_file.write("{} {}\n".format(SNAME, VERSION))
ver_file.close()


DATA_VERSION = '1.0'
SHARED_MEMORY_PATH = "/dev/shm/control.json"
TEMP_PATH = "/dev/shm/control.json.tmp"

def now():
    return datetime.datetime.now(datetime.UTC).isoformat()


class LogWriter(object):

    def __init__(self, filename):
        self.filename = filename
        self.clear()

    def clear(self):
        self.overwrite_last = False
        self.data = {'version': DATA_VERSION,
                     'start': now(), 'messages': [], 'status': 'new test'}
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
            self.end()
        self._write()

    def changelast(self, s):
        self.data['messages'][-1] = s
        self._write()

    # `def _write(self):        
    #     with open(self.filename + ".tmp", "w") as f:
    #         json.dump(self.data, f, indent=2)
    #         os.replace(self.filename + ".tmp", self.filename) # This updates the
    #         # file atomically, which can be important
    def _write(self):
        with open(TEMP_PATH, "w") as f:
            json.dump(data, f, indent=2)
        # Atomic swap within RAM
        os.replace(TEMP_PATH, SHARED_MEMORY_PATH)

    def setserial(self, s):
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


currentdirectory = os.path.dirname(__file__)
thefile = os.path.join(currentdirectory, "..",  "control.json")
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
