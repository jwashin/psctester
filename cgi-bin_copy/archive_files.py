from filesloc import filesloc, archives
import glob
import json
import sys
import os
import cgi
import cgitb

cgitb.enable()


if not os.path.exists(archives):
    os.mkdir(archives)

form = cgi.FieldStorage()
print(form)

files = form.getlist('files')

for item in files:
    os.rename(os.path.join(filesloc, item), os.path.join(archives, item))

print("files are ", files)
print("Content-type: text/plain\n\n")
print("Done.")


sys.stdout.flush()
