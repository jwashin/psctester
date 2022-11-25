VERSION = 2.0
SNAME = "ACsupply.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

#import hardware_interface

class ACsupply_obj():

    def __init__(self):
        self.specified_amps = 0.0
        self.specified_volts = 0.0
        return
    
    def set_volts(self,  pass_volts=120):
        self.specified_volts = pass_volts
        if pass_volts > 0:
            print("volts off")
        else:
            print("volts on")

        return True

    def set_amps(self,  pass_amps):
        self.specified_amps = pass_amps
        if pass_amps <= 0:
            print("amps off")
        elif pass_amps <= 10:
            print("amps on low")
        else:
            print("amps on high")

        return True

    def amps_on():
        print("Amps on!")
        return True

    def volts_on():
        print("Volts on!")
        return True

    def amps_off():
        print("Amps off!")
        return True

    def volts_off():
        print("Volts off!")
        return True

    
    
                  
