import cgi
import sys
import os

# for debugging
#import cgitb
#cgitb.enable()

from filesloc import filesloc

form = cgi.FieldStorage()

filename = form['filename'].value

if filename:
    os.chdir(filesloc)
    print ('Content-Type: text/csv\nContent-Disposition: attachment; filename="{}"\n'.format(filename))
    with open(filename, 'r') as testfile:
        z = testfile.readlines()
    for k in z:
        print ("{}".format(k.strip()))
    sys.stdout.flush()
 

