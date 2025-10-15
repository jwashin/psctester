import os
filesloc = "/home/pi/acc_tester/reports"
archives = "/home/pi/acc_tester/archive"
if not os.getenv("USER") == 'root':
    home = os.getenv('HOME')
    filesloc = os.path.join(home, "acc_tester", "reports")
    archives = os.path.join(home, "acc_tester", "archive")
if not os.path.exists(filesloc):
    os.makedirs(filesloc, exist_ok=True)
if not os.path.exists(archives):
    os.makedirs(archives, exist_ok=True)
