import logwriter
import sys

def initfile():

    lw = logwriter.logwriter
    lw.clear()

initfile()

print("Content-type: text/plain")
print("")
sys.stdout.flush()
