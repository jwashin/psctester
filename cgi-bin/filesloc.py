import os
filesloc = "/home/pi/acc_tester/reports"
archives = "/home/pi/acc_tester/archive"
if not os.path.exists(filesloc):
    os.mkdir(filesloc)
if not os.path.exists(archives):
    os.mkdir(archives)
    
