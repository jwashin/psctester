##VERSION = 2.0
VERSION = 2.01  # Improve manual control when run as main

SNAME = "radian.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

import serial

import struct
import logging


Q_logger = logging.getLogger('QLC_acc_tstr')

class Radian_obj(serial.Serial):
    
#''' Version 1.0   29 sep 2014'''

    class Radian_packet:

        def __init__(self):
            self.p_start = b'\xa6'
            self.p_type = b'\x00'
            self.p_length = 0
            self.p_data = b''
            self.p_checksum = 0
            return

        def cheksum(buf):
            tot = 0
            for c in buf:
                tot += c
            return tot

        def format_buf(self):

            self.c_buf = struct.pack(">ssH",
                                      self.p_start,
                                      self.p_type,
                                      self.p_length)
            
            if self.p_length:
                self.c_buf += struct.pack(str(self.p_length)+"s",
                                      self.p_data)
                
            self.c_buf += struct.pack(">H",
                                      Radian_obj.Radian_packet.cheksum(self.c_buf))
                                      
            return

        def process_rbuf(self):
            ret_code = False                # preset return code to failed
            if len(self.r_buf) < 2:
                pass
#                print ("No Reply")
#                errCode = 0
            else:
                self.p_checksum = struct.unpack(">H",self.r_buf[-2:])[0]
                if self.p_checksum == Radian_obj.Radian_packet.cheksum(self.r_buf[:-2]):
                    (self.p_start, self.p_type, self.p_length) = struct.unpack(">ssH",self.r_buf[:4])
                    if self.p_start == b'\a9':  # NOTE - NAK processing not complete!!!
                        pass
#                        errCode = 1
#                        print("NAK")
                    else:
                        ret_code = True
                    if self.p_length:
                        self.p_data = self.r_buf[4:-2]
                    # deal with checksum error here!!!!!!!!!!!!!!!!!!
                    # also check for NAk and p_type mismatch !!!!!
                #print (self.p_start,self.p_type,self.p_length,self.p_data)
            return ret_code

    def __init__(self,ttydev):
        self.ttydev = ttydev
        super().__init__(self.ttydev,            # Serial port parameters for Radian RS-232
                    baudrate=57600,
                    bytesize =8,
                    parity='N',
                    stopbits=1,
                    timeout = 0.5)

        self.__out_packet__ = Radian_obj.Radian_packet()
        self.__in_packet__ = Radian_obj.Radian_packet()
        self.send_cmd("Identification")
        self.send_cmd("Read Calibration",b'\x00\x08\xA7\x00\x4F')
        
        return

    def __nd_string__(self):
        ret_str = ""
        for i in range(self.nd_ptr,self.__in_packet__.p_length):
            w = self.__in_packet__.p_data[i:i+1]
            if w == b',':
                break
            elif w != b'\x00':
                ret_str += chr(w[0])

        self.nd_ptr = i+1
        return ret_str

    def __ti_float__(self,pass_int):
        expon = (pass_int >> 24) & 0xff
        if (expon == 0x80) or (expon == 0x81):
            return 0.0
        if expon > 127:
            expon = expon - 256

        mant = pass_int & 0x7fffff
        
        if pass_int & 0x800000:
            mant = mant - (2 ** 24)
        else:
            mant = mant | 0x800000

        #print("exponent = {0:x},mantissa = {1:x}, fraction = {2}".format(expon,mant,mant/2**23))
        
        return (mant / (2 ** 23)) * (2 ** expon)

    def send_cmd(self,msg_type,msg_data=b''):
        
        self.__out_packet__.p_type = p_code_from_name[msg_type]
        self.__out_packet__.p_length = len(msg_data)
        self.__out_packet__.p_data = msg_data
        self.__out_packet__.format_buf()
        cmd_try_count = 3
        while cmd_try_count:
            
            self.write(self.__out_packet__.c_buf)
##            print('send\n{}'.format(hexprint(self.__out_packet__.c_buf)))
            self.__in_packet__.r_buf = self.read(400)
##            print('receive\n{}'.format(hexprint(self.__in_packet__.r_buf)))
            #  Add some error checking here!!!!!!!
            if self.__in_packet__.process_rbuf():
                cmd_try_count = False
            else:
                cmd_try_count -= 1

        if self.__in_packet__.p_type == b'\x02':
            self.nd_ptr = 0
            self.model = self.__nd_string__()
            self.serial = self.__nd_string__()
            self.name = self.__nd_string__()
            self.version = self.__nd_string__()
            Q_logger.info('Radian version: {},{},{},{}'.format(self.model, self.serial, self.name, self.version))
        elif self.__in_packet__.p_type == b'\x04':
            self.cal_date = (self.__in_packet__.p_data[4:6].decode("utf-8") + "/" + 
                                self.__in_packet__.p_data[6:8].decode("utf-8") + "/" + 
                                self.__in_packet__.p_data[0:4].decode("utf-8"))
            Q_logger.info('Radian cal date: {}'.format(self.cal_date))
        elif self.__in_packet__.p_type == b'\x0d':
            h = struct.unpack(">9I",self.__in_packet__.p_data)
            self.volts = self.__ti_float__(h[0])
            self.amps = self.__ti_float__(h[1])
            self.watts = self.__ti_float__(h[2])
            self.volt_amps = self.__ti_float__(h[3])
            self.VARs = self.__ti_float__(h[4])
            self.frequency = self.__ti_float__(h[5])
            self.degrees_phase = self.__ti_float__(h[6])
            self.power_factor = self.__ti_float__(h[7])
            self.delta_phase = self.__ti_float__(h[8])
            Q_logger.info('Radian Inst. values:  {:.4g}V {:.4g}A {:.4g}W {:.4g}VA {:.4g}VAR {:.4g}Hz {:.4g}deg {:.4g}PF {:.4g}deg'.format(self.volts, 
                                                                                                      self.amps, 
                                                                                                      self.watts,
                                                                                                      self.volt_amps, 
                                                                                                      self.VARs, 
                                                                                                      self.frequency, 
                                                                                                      self.degrees_phase, 
                                                                                                      self.power_factor, 
                                                                                                      self.delta_phase))
        elif self.__in_packet__.p_type == b'\x16':
            h = struct.unpack(">8I",self.__in_packet__.p_data)
            self.watt_hours = self.__ti_float__(h[0])
            self.VAR_hours = self.__ti_float__(h[1])
            self.Q_metric_hours = self.__ti_float__(h[2])
            self.volt_amp_hours = self.__ti_float__(h[3])
            self.volt_hours = self.__ti_float__(h[4])
            self.amp_hours = self.__ti_float__(h[5])
            self.volt_squared_hours = self.__ti_float__(h[6])
            self.amp_squared_hours = self.__ti_float__(h[7])
            Q_logger.info('Radian Acc. values:  {:.4g}WH {:.4g}VARH {:.4g}QH {:.4g}VAH {:.4g}VH {:.4g}AH {:.4g}V2H {:.4g}A2H'.format(self.watt_hours, 
                                                                                                     self.VAR_hours, 
                                                                                                     self.Q_metric_hours, 
                                                                                                     self.volt_amp_hours, 
                                                                                                     self.volt_hours, 
                                                                                                     self.amp_hours, 
                                                                                                     self.volt_squared_hours, 
                                                                                                     self.amp_squared_hours))
        elif p_name_from_code[self.__in_packet__.p_type][1] & 0x01:  # shorted out!!!
            pass
#            errCode = 2
#            print("Warning - no handler for receive packet type {}".format(self.__in_packet__.p_type))
            
            
                
            
            
        return

    def update_readings(self):
        """ Read instantaneous values from Radian """
        
        self.send_cmd("Read Instantaneous Metrics RD-2x format",b'\x00\x24\x00\x00\x00\x14\xff\xfd')

        return

    def setup_for_gate(self):
        """ Set up Radian to respond to gate pulse """

        self.send_cmd("Stop Accumulating Metrics")   #stop accumulation

        self.send_cmd("BNC Control",b'\x01\x01')   # set up BNC 1 for start-and-clear/stop
        
        self.send_cmd("Start Accumulating Metrics",b'\x06\x00\x01')   # setup to start when button pushed

        return

    def read_accumulated(self):

        self.send_cmd("Read Accumulated Metrics RD-2x format",b'\x00\x20\x00\x00\x00\x04')   # read metrics
        return
    

        

# Dictionary used for packet processing.
# Key = Packet Id in hex
# Value = 3-tuple with the following elements:
# [0] = Name of command (string)
# [1] = method to call to format the send packet
# [2] = method to call to handle the receive packet
#   future!!!!!!!!

# key = code byte in radian packet for particular packet type"
# data = (packet name, packet flags)\
# packet flags = Bit 7-1: reserved
#                Bit 0: set to indicate receive packet needs unpacking of data
#
p_name_from_code = {
    b'\x00' : ("NOP",0),
    b'\x02' : ("Identification",1),
    b'\x03' : ("Reset RD",0),
    b'\x04' : ("Read Calibration",1),
    b'\x07' : ("Reset Metrics",0),
    b'\x08' : ("Start Accumulating Metrics",0),
    b'\x09' : ("Stop Accumulating Metrics",0),
    b'\x0a' : ("Start A Timed Accumulating Test",0),
    b'\x0b' : ("Lock/Unlock Relay Ranges",0),
    b'\x0c' : ("Trigger Waveform",0),
    b'\x0d' : ("Read Instantaneous Metrics RD-2x format",1),
    b'\x0e' : ("Accumulated Waveform Data Read",1),
    b'\x0f' : ("Harmonic Data Read",1),
    b'\x16' : ("Read Accumulated Metrics RD-2x format",1),
    b'\x1b' : ("Auto-Calibrate",0),
    b'\x1d' : ("BNC Control",0),
    b'\x20' : ("System Status",1),
    b'\x21' : ("Minimum Metrics Data Read RD-2x format",1),
    b'\x23' : ("Maximum Metrics Data Read RD-2x format",1),
    b'\x28' : ("Trigger Harmonic Analysis",1),
    b'\x2c' : ("Mode Change",0),
    b'\x2e' : ("Read Instantaneous Metrics RD-3x format",1),
    b'\x2f' : ("Read Accumulated Metrics RD-3x format",1),
    b'\x30' : ("Minimum Metrics Data Read RD-3x format",1),
    b'\x31' : ("Maximum Metrics Data Read RD-3x format",1),
    b'\x32' : ("Pulse Output Constant Change",0),
    b'\x34' : ("Standard Test",1),
    b'\x39' : ("Meter Test",1)
    }

p_code_from_name = {}

for i in p_name_from_code:
    p_code_from_name[p_name_from_code[i][0]] = i 

def hexprint(buf):
    return ":".join("{:02x}".format(c) for c in buf)


if __name__ == "__main__":

    Radian = Radian_obj("/dev/tty_radian")
    
    print(Radian.model,Radian.serial,Radian.name,Radian.version)

    print("Last calibrated: {}".format(Radian.cal_date))

    Radian.update_readings()
            
    print("""Volts = {:8.3f}, Amps = {:8.3g}, Watts = {:8.3g}, VA = {:8.3g}, VAR = {:8.3g}, Frequency = {:8.3f}
Degrees Phase = {:8.3f}, PF = {:8.3f}, Delta Phase = {:8.3g}""".format(Radian.volts,
                                                                       Radian.amps,
                                                                       Radian.watts,
                                                                       Radian.volt_amps,
                                                                       Radian.VARs,
                                                                       Radian.frequency,
                                                                       Radian.degrees_phase,
                                                                       Radian.power_factor,
                                                                       Radian.delta_phase
                                                                       ))

    Radian.send_cmd("Read Calibration",b'\x00\x08\xA7\x00\x4F')

    print(Radian.cal_date)

    Radian.send_cmd("Accumulated Waveform Data Read",b'\x00\x08\x00\x00\x00\x04')   # should cause warning
    
    Radian.setup_for_gate()   # setup to start when button pushed

    while True:
        user_input = input('Enter command (r = reset Radian, q = quit, anything else = display KWH)--->')
        if user_input == 'q':
            break
        elif user_input == 'r':
            Radian.setup_for_gate()
        else:
            Radian.read_accumulated()   # read metrics
            print(Radian.watt_hours)
        
    

    

    



    
