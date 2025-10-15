import glob
import json
import sys
import os

from filesloc import filesloc

os.chdir(filesloc)

search_dir = "."
# remove anything from the list that is not a file (directories, symlinks)
# thanks to J.F. Sebastion for pointing out that the requirement was a list 
# of files (presumably not including directories)  
files = glob.glob("*.csv")
files.sort(key=lambda x: os.path.getmtime(x))
files.reverse()

print ("Content-type: text/json\n\n")

print (json.dumps({'files':files}))

sys.stdout.flush()

