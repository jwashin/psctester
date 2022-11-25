#VERSION = 2.0
#VERSION = 2.01       # Add logging
#VERSION = 2.02       # Fix bugs in logging
##VERSION = 2.03      # Update manual control code
VERSION = 2.04      # Add Start Radian to manual control code

SNAME = "hardware_interface.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

import time
import RPi.GPIO as GPIO
import logging

AMPS_ON_RELAY = 22                      # Set to 0 to turn current loop on
AMPS_HIGH_RELAY = 11                    # Set to 0 to short out resistor, giving high current
AMPS_LOW_RELAY = 16                     # set to 0 to turn on primary current
VOLTS_ON_RELAY = 15                     # set to 0 to turn on voltage

RELAY_DELAY = 0.1                       # Seconds to wait after switching relay

GATE_CONTROL = 12                       # pulse low for 50mS to start/stop gating
SUPPRESS_OPTO_GATE = 32                 # 1 = allow opto gate pulses, 0 = suppress

STATUS_LED = 18                         # 0 = RED, 1 = GREEN

OVER_TEMP = 31                          # Flag which indicates overheat indication when pulled low

Q_logger = logging.getLogger('QLC_acc_tstr')

class Tester_hardware_obj():

    def __init__(self):

        GPIO.setmode(GPIO.BOARD)
        
        GPIO.setup(AMPS_ON_RELAY, GPIO.OUT, initial=1)
        
        GPIO.setup(AMPS_HIGH_RELAY, GPIO.OUT, initial=1)
        
        GPIO.setup(AMPS_LOW_RELAY, GPIO.OUT, initial=1)
        
        GPIO.setup(VOLTS_ON_RELAY, GPIO.OUT, initial=1)
        
        GPIO.setup(GATE_CONTROL, GPIO.OUT, initial=1)
        
        GPIO.setup(SUPPRESS_OPTO_GATE, GPIO.OUT, initial=1)

        GPIO.setup(OVER_TEMP, GPIO.IN, pull_up_down=GPIO.PUD_UP)

#        GPIO.setup(STATUS_LED, GPIO.OUT, initial=1)
        

        return

    def set_amps_off(self):
        Q_logger.info('Turn current off')
        GPIO.output(AMPS_ON_RELAY,1)
        return

    def set_amps_on(self):
        Q_logger.info('Turn current on')
        GPIO.output(AMPS_ON_RELAY,0)
        return

    def preset_amps_high(self):
        self.set_amps_off()
        time.sleep(RELAY_DELAY)
        Q_logger.info('Set up for high current')
        GPIO.output(AMPS_LOW_RELAY,0)
        GPIO.output(AMPS_HIGH_RELAY,0)
        time.sleep(RELAY_DELAY)
        return

    def preset_amps_low(self):
        self.set_amps_off()
        time.sleep(RELAY_DELAY)
        Q_logger.info('Set up for low current')
        GPIO.output(AMPS_LOW_RELAY,0)
        GPIO.output(AMPS_HIGH_RELAY,1)
        time.sleep(RELAY_DELAY)
        return

    def set_volts_off(self):
        Q_logger.info('Turn voltage off')
        GPIO.output(VOLTS_ON_RELAY,1)
        return

    def set_volts_on(self):
        Q_logger.info('Turn voltage on')
        GPIO.output(VOLTS_ON_RELAY,0)
        return

    def do_gate_pulse(self, time_ms=50):
        Q_logger.info('Send gate pulse to Radian')
        GPIO.output(GATE_CONTROL,0)
        time.sleep(time_ms/1000)
        GPIO.output(GATE_CONTROL,1)
        return

    def suppress_opto_gate(self):
        Q_logger.info('Disable opto gate pulses')
        GPIO.output(SUPPRESS_OPTO_GATE,0)
        return

    def allow_opto_gate(self):
        Q_logger.info('Enable opto gate pulses')
        GPIO.output(SUPPRESS_OPTO_GATE,1)
        return

    def do_cleanup(self):
        GPIO.cleanup()
        return
        
    
if __name__ == "__main__":
    
#    import meter
    
    print("start")


    tester = Tester_hardware_obj()

    keep_running = True

    def exit_hw_interface():
        global keep_running
        keep_running = False

    def set_amps_low():
        tester.preset_amps_low()
        tester.set_amps_on()

    def set_amps_high():
        tester.preset_amps_high()
        tester.set_amps_on()
        

    
    options = {"amps off":tester.set_amps_off,
               "volts off":tester.set_volts_off,
               "volts on":tester.set_volts_on,
               "amps on low":set_amps_low,
               "amps off":tester.set_amps_off,
               "amps on high":set_amps_high,
               "Start Radian accumulation" : tester.do_gate_pulse, 
               "exit":exit_hw_interface}

    print("Choose one of the following options")
    
    print(*options.keys(), sep='\n' )

    while keep_running:

        choice = input("--> ")
        
        if choice in options:
            options[choice]()
        else:
            print("not a valid option, please choose one of the following")
            for valid_command in options.keys():
                print(valid_command)

    

###    DUT = meter.Meter_obj("/dev/ttyUSB1")
##
####    tester.set_amps_off()
####
####    tester.set_amps_high()
##
##    input("volts on")
##    tester.set_volts_on()
##    cont = True
##    while cont:
##        
##        if (cont and len(input("amps off"))==0):
##            tester.set_amps_off()
##        else:
##            cont = False
##            
##        if (cont and len(input("amps on low"))==0):
##            tester.preset_amps_low()
##            tester.set_amps_on()
##        else:
##            cont = False
##            
##        if (cont and len(input("amps off"))==0):
##            tester.set_amps_off()
##        else:
##            cont = False
##            
##        if (cont and len(input("amps on high"))==0):
##            tester.preset_amps_high()
##            tester.set_amps_on()
##        else:
##            cont = False
##                    
###        time.sleep(1)
##        if GPIO.input(OVER_TEMP):
##            print('cool')
##        else:
##            print('hot',b'\x07'.decode('utf-8'))
##
##    if len(input("volts off"))==0:
##        tester.set_volts_off()
##    else:
##        cont = False
                    

        
##        tester.set_amps_off()
        
##        tester.do_gate_pulse()
##        time.sleep(1)
###        print("IN LEN",DUT.com.inWaiting())
##        DUT.com.flushInput()
##        print("flush LEN",DUT.com.inWaiting())
##        rd_char = DUT.com.read(8000)
##        print("1:",len(rd_char),rd_char)
##        print("IN LEN",DUT.com.inWaiting())
##        rd_char = DUT.com.read(8000)
##        print("2:",len(rd_char),rd_char)


##        tester.do_gate_pulse()
##        time.sleep(0.1)
      
    #GPIO.cleanup()
    tester.do_cleanup()
    
    
