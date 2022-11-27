import cgi
import json
import zipfile
import sys
import os
#import cgitb
#cgitb.enable()

from filesloc import filesloc
os.chdir(filesloc)

form = cgi.FieldStorage()

files = form.getlist('files')


myfile = zipfile.ZipFile('psc_zip.zip', mode='w')
for item in files:
    myfile.write(item, compress_type=zipfile.ZIP_DEFLATED)

myfile.close()

fn = sys.stdout.fileno()

print ('Content-Type: application/zip\nContent-Disposition: attachment; filename="psc_zip.zip"\n')

#print ("{}".format(json.dumps({'resp':files})))
sys.stdout.flush()

f = open("psc_zip.zip", 'rb')

while True:
    t = f.read()
    if not t:
        break
    os.write(fn, t)
