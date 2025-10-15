#!/usr/bin/python3 -u

VERSION = 2.01          # Add Q Family support to PSC tester

SNAME = "PSC_QFAM_interface.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()


import os
import time
import calendar

import wiringpi

from pymodbus.constants import Endian

from pymodbus.payload import BinaryPayloadDecoder

from pymodbus.client import ModbusSerialClient as ModbusClient

from QLC_mb_addr_set import set_modbus_addr_obj


#
#   Global variables
#

i2c_obj = wiringpi.I2C()         # Set up i2c interface global

debug_on = False


import QFAM_equates #import CHANNELS_PER_METER, NUM_METERS, READ_VERSION, I2C_RBUF_LEN, CMD_BUSY, DATA_LENGTH, READ_TEST_DATA, READ_LATEST_DATA, READ_TEST_STATUS, START_TEST, WORK_MODBUS_ADDR



#
#   Class definitions
#

class i2c_buf():
    pass

class PulseCnt():
    def __init__(self,  mtr = 0, rad = 0):
        self.meter = mtr
        self.radian = rad
        return
        
class PhaseData():
    def __init__(self,  chan_name = '',  point_name = ''):
        self.chan_name = chan_name
        self.point_name = point_name
        self.volts = None
        self.amps = None
        self.watts = None
        self.vars = None
        self.va = None
        self.pf = None
        self.freq = None
        self.kwh_del = None
        self.kwh_rcv = None
        self.kvarh_1 = None
        self.kvarh_2 = None
        self.kvarh_3 = None
        self.kvarh_4 = None
        self.kvah_del = None
        self.kvah_rcv = None
        return
        
class PointData():
    def __init__(self):
        self.chan = [PhaseData() for i in range(QFAM_equates.CHANNELS_PER_METER)]
        self.tot = PhaseData()
        return
        
class MeterData():
    def __init__(self):
        self.modbus_addr = None
        self.serno = None
        self.version = None
        self.pulse_type = None
        self.date_and_time= None
        self.pulse_weight = None
        self.comm_OK = False
        self.wiring_mode = None
        self.voltage_range = None
        self.ct_type = None
        self.pt_ratio = None
        self.primary_amp = [None for i in range(QFAM_equates.NUM_METERS)]
        self.point = [PointData() for i in range(QFAM_equates.NUM_METERS)]
        return

class PulseTest():
    def __init__(self, i2c_addr, tty_port):
        self.tty_port = tty_port
        self.port_OK = False
        self.i2c_addr = i2c_addr
        self.meter = MeterData()
        self.modbus_client = ModbusClient(method='rtu',
                             port=self.tty_port,
                             timeout=1,
                             broadcast_enable = True, 
                             baudrate=9600)
                             
        self.modbus_masks = set_modbus_addr_obj(self.modbus_client)
        self.set_masks = True
        self.msp430_OK = self.i2c_setup()
        self.msp430_version = None
        self.msp430_serno = None
        self.last_msp430_refresh = time.time()
        self.first = [0] * QFAM_equates.NUM_METERS
        self.latest = [PulseCnt() for i in range(QFAM_equates.NUM_METERS)]
        self.latest_Radian_count = 0
        self.final = [PulseCnt()for i in range (QFAM_equates.NUM_METERS)]
        self.current = [PulseCnt()for i in range (QFAM_equates.NUM_METERS)]
        self.test_status = bytearray(QFAM_equates.NUM_METERS)
        self.sBuf = i2c_buf()
        self.sBuf.cmd = 0
        self.sBuf.params = []
        self.rBuf = i2c_buf()
        self.rBuf.status = 0
        self.rBuf.data_ID = 0
        self.rBuf.params = []
        self.rBuf.bytes = bytearray()
        self.get_version()
 
        return
        
    def i2c_setup(self):
        self.msp430_OK = False
        self.i2c_handle = i2c_obj.setup(self.i2c_addr)
        if self.i2c_handle != -1:
            try:
                os.write(self.i2c_handle, b'0')
                self.msp430_OK = True
            except Exception:
                try:
                    os.close(self.i2c_handle)
                except Exception:
                    pass
                pass
        return (self.msp430_OK)



    def i2c_exec(self, cmd = QFAM_equates.READ_VERSION, parms = []):
        ret_code = False
        self.sBuf.cmd = cmd
        self.sBuf.params = parms
        self.last_msp430_refresh = time.time()
            
        if (not self.msp430_OK) and (not self.i2c_setup()):
            return (False)
        s_data = bytearray(QFAM_equates.I2C_SBUF_LEN)
        s_data[0] = self.sBuf.cmd << 4
        s_index = 1
        for s_int in self.sBuf.params:
            s_data[s_index:s_index+2] = s_int.to_bytes(2,byteorder='little')
            s_index += 2
        work_checksum = 0
        for i in range(0,s_index):
            work_checksum += s_data[i]
            
        s_data[s_index:s_index+2] = work_checksum.to_bytes(2,byteorder='little')
        
        if debug_on:
            print('snd: ',' '.join('{:02x}'.format(x) for x in s_data[0:s_index+2]))

        try:
            count = os.write(self.i2c_handle, s_data[0:s_index+2])
        except Exception:
            self.msp430_OK = False
            return (False)
        
        if (count != s_index+2):
            self.MSP_430_OK = False
            return (False)
        
        MSP430_done = False
        MSP430_retries = 5
        while ((not MSP430_done) and (MSP430_retries > 0)):
            try:
                reply_bytes = os.read(self.i2c_handle,QFAM_equates.I2C_RBUF_LEN)
            except Exception:
                self.msp430_OK = False
                return (False)
            if debug_on:
                print('rcv: ',' '.join('{:02x}'.format(x) for x in reply_bytes))
            self.rBuf.status = reply_bytes[0]
            MSP430_done = not (self.rBuf.status & QFAM_equates.CMD_BUSY)
            MSP430_retries -= 1
            
        self.rBuf.data_ID = reply_bytes[1]
        r_data_len = QFAM_equates.DATA_LENGTH[self.rBuf.data_ID]
        self.rBuf.bytes = reply_bytes[2 : r_data_len + 2]
        r_index = 2
        self.rBuf.params = []
        while r_index < r_data_len + 2:
            self.rBuf.params += [int.from_bytes(reply_bytes[r_index:r_index+2],byteorder='little')]
            r_index += 2

        work_checksum = 0
        for i in range(0,r_index):
            work_checksum += reply_bytes[i]
            
        ret_code = reply_bytes[r_index:r_index+2] == work_checksum.to_bytes(2,byteorder='little')

        if debug_on:
            print('i2c return status: {}'.format(ret_code))
        
        return(ret_code)
        
    def read_regs(self, start_reg,  num_regs):

        if self.set_masks or not self.port_OK:
            self.set_masks = False
            try:
                self.modbus_masks.send(action = 1, 
                                        timeout = 1, 
                                        true_addr = QFAM_equates.WORK_MODBUS_ADDR, 
                                        not_addr = 0, 
                                        match_addr = 0, 
                                        and_mask = 0x00000000, 
                                        or_mask = 0xFFFFFFFF)       # Set modbus address to 1
            except Exception as e:
                if debug_on:
                    print('Exception while sending masks: port = {}, {}'.format(self.modbus_client.port, e))
                self.port_OK = False
                self.meter.comm_OK = False
                return (None)
                
            self.port_OK = True
        try:
            result = self.modbus_client.read_holding_registers(start_reg,  num_regs,  QFAM_equates.WORK_MODBUS_ADDR)
        except Exception as e:
            if debug_on:
                print('Exception while reading registers: port = {}, {}'.format(self.modbus_client.port, e))
            result = None

        if result:
            if result.isError():
                if debug_on:
                    print('{}'.format(result))
                result = None
                
        if result == None:
            self.meter.comm_OK = False
            return (None)            
        else:
            self.meter.comm_OK = True
            return (result.registers)
            
    def comm_err_msg(self):
        if not self.port_OK:
            return ('port offline')
        if not self.meter.comm_OK:
            return ('meter offline')
        return('unknown')

    def get_version(self):
        if not self.i2c_exec():
            self.msp430_version = None
            self.msp430_serno = None
        else:
            self.msp430_version = '{}.{}'.format(self.rBuf.params[0] % 256, self.rBuf.params[0] // 256)
            self.msp430_serno = self.rBuf.params[1]
        return
        
    def get_test_data(self):
        if not self.i2c_exec(QFAM_equates.READ_TEST_DATA):
            self.current = [PulseCnt() for i in range (QFAM_equates.NUM_METERS)]
        else:
            for i in range(len(self.current)):
                self.current[i] = PulseCnt(int.from_bytes(self.rBuf.bytes[i * 6 : i * 6 + 2], byteorder='little'),             
                                    int.from_bytes(self.rBuf.bytes[i * 6 + 2 : i * 6 + 6], byteorder='little'))
        return
        
    def get_latest_data(self):
        if not self.i2c_exec(QFAM_equates.READ_LATEST_DATA):
            self.latest = [PulseCnt() for i in range(QFAM_equates.NUM_METERS)]
        else:
            for i in range(len(self.latest)):
                self.latest[i] = PulseCnt(int.from_bytes(self.rBuf.bytes[i * 6 : i * 6 + 2], byteorder='little'),             
                                    int.from_bytes(self.rBuf.bytes[i * 6 + 2 : i * 6 + 6], byteorder='little'))
        return
        
    def get_test_status(self):
        if not self.i2c_exec(QFAM_equates.READ_TEST_STATUS):
            self.latest_Radian_count = 0
            self.first = [0] * QFAM_equates.NUM_METERS
        else:
            self.latest_Radian_count = self.rBuf.params[0] + (self.rBuf.params[1] << 16)
            self.test_status = self.rBuf.bytes[4:10]
            for i in range(len(self.first)):
                self.first[i] = int.from_bytes(self.rBuf.bytes[i * 4 + 10:i * 4 + 14], byteorder='little')            
        return
        
    def start_test(self, meter_pulses):
        self.i2c_exec(QFAM_equates.START_TEST, [meter_pulses])
        return
        
    def get_meter_info(self):
        
        self.meter.serno = None
        self.meter.version = None
        self.meter.pulse_weight = None
        self.meter.pulse_type = None
        self.meter.date_and_time = None
        self.meter.modbus_addr = None
        self.meter.wiring_mode = None
        self.meter.voltage_range = None
        self.meter.ct_type = None
        self.meter.pt_ratio = None
        self.meter.primary_amp = [None for i in range(QFAM_equates.NUM_METERS)]

#         #      NOTE: test only!!!!!!!!
#
#        self.start_test(32000)
#
#        while True:
#            self.get_test_status()
#            self.get_latest_data()
#            print(self.test_status, "Radian = ",  self.latest_Radian_count)
#            for i in range(len(self.latest)):
#                print(i,  self.latest[i].meter)
#                pass
#                
#                # end of test section
                
        result =  self.read_regs(1330, 3)
        if result == None:
            return (False)
        else:
            self.meter.serno = (result[2] << 16) + result[1]
            self.meter.version = result[0]
            
        result =  self.read_regs(1278, 2)
        if result == None:
            return (False)
        else:
            self.meter.pulse_weight = result[0]
            self.meter.pulse_type = {
                                0 : 'WH', 
                                1 : 'VARH', 
                                2 : 'VAH'
                                } [result[1]]
            
        result =  self.read_regs(1270, 4)
        if result == None:
            return (False)
        else:

            self.meter.date_and_time= '{:02d}/{:02d}/{}{:02d} {:02d}:{:02d}:{:02d} {}'.format(
                                                            result[0] & 0xff,                                                       # LSB = month
                                                            result[1] >> 8,                                                         # MSB = day of month
                                                            QFAM_equates.DEFAULT_CENTURY_PREFIX,             # Convert 2-digit to 4-digit year
                                                            result[0] >> 8,                                                         # LSB = 2-digit year
                                                            result[1] & 0xff,                                                       # LSB = hour
                                                            result[2] >> 8,                                                         # MSB = minute
                                                            result[2] & 0xff,                                                       # LSB = second
                                                            calendar.day_name[(result[3] >> 8) - 1]
                                                            )

        result =  self.read_regs(1210, 1)
        if result == None:
            return (False)
        else:
            self.meter.modbus_addr = result[0]
            
        result =  self.read_regs(1240, 10)
        if result == None:
            return (False)
        else:
            (self.meter.wiring_mode, self.meter.voltage_range,  self.meter.ct_type ,  self.meter.pt_ratio) = result[0:4]
            self.meter.primary_amp = result[4:]
            return (True)
            
    def get_phase_diagnostics(self, mtr = None, chn = None, request = 'all'):
        
        def fl(reg_list):
            ret_list = []
            for i in range(0, len(reg_list), 2):
                ret_list += [BinaryPayloadDecoder.fromRegisters(reg_list[i:i+2], Endian.Big, wordorder=Endian.Little).decode_32bit_float()]
            return (ret_list)    
                
        def lg(reg_list):
            ret_list = []
            for i in range(0, len(reg_list), 2):
                ret_list += [(reg_list[i] + (reg_list[i+1] << 16)) / 1000]
            return (ret_list)    
                
        
        def get_pd_data(point, channel=None):       # Channel = None means total
        
            if request not in ['all', 'inst', 'acc']:
                return (False)
            if channel == None:
                c = PhaseData(chan_name = 'tot')
                self.meter.point[point].tot = c
            else:
                c = PhaseData(chan_name = str(channel + 1))
                self.meter.point[point].chan[channel] = c
            c.point_name = str(point + 1)
            r = point * 150             #  Start of modbus regs for this metering point

            if request in ['all', 'inst']:
                if channel != None:
                    o = 0 + 2 * channel
                    result =  self.read_regs(r + o, 2)
                    if result == None:
                        return (False)
                    else:
                        c.volts = fl(result)[0]
                
                if channel != None:
                    o = 12 + 2 * channel
                    result =  self.read_regs(r + o, 2)
                    if result == None:
                        return (False)
                    else:
                        c.amps = fl(result)[0]
                
                if channel != None:
                    o = 18 + 2 * channel
                else:
                    o = 24
                result =  self.read_regs(r + o, 2)
                if result == None:
                    return (False)
                else:
                    c.watts = fl(result)[0]
                
                if channel != None:
                    o = 26 + 2 * channel
                else:
                    o = 32
                result =  self.read_regs(r + o, 2)
                if result == None:
                    return (False)
                else:
                    c.vars = fl(result)[0]
                
                if channel != None:
                    o = 34 + 2 * channel
                else:
                    o = 40
                result =  self.read_regs(r + o, 2)
                if result == None:
                    return (False)
                else:
                    c.va = fl(result)[0]
                
                if channel != None:
                    o = 42 + 2 * channel
                else:
                    o = 48
                result =  self.read_regs(r + o, 2)
                if result == None:
                    return (False)
                else:
                    c.pf = fl(result)[0]
                    
                o = 50
                result =  self.read_regs(r + o, 2)
                if result == None:
                    return (False)
                else:
                    c.freq = fl(result)[0]
                    
            if request in ['all', 'acc']:
                
                if channel != None:
                    o = 52 + 16 * channel
                else:
                    o = 100
                result =  self.read_regs(r + o, 16)
                if result == None:
                    return (False)
                else:
                    (c.kwh_del,   
                     c.kwh_rcv,  
                     c.kvarh_1,  
                     c.kvarh_2,  
                     c.kvarh_3,  
                     c.kvarh_4,  
                     c.kvah_del, 
                     c.kvah_rcv ) = lg(result)
                
                
            
            return (True)
            
        for point in range(len(self.meter.point)):
            if (mtr == None) or (mtr == point):
                if not  get_pd_data(point):
                    return(False)
                for chan in range(len(self.meter.point[point].chan)):
                    if (chn == None) or (chn == chan):
                        if not  get_pd_data(point, chan):
                            return (False)
        return (True)
                
    def set_meter_pulses(self, pulse_type = 'WH', pulse_weight = 1000):
        
        try:
            
            self.modbus_masks.send(action = 1, 
                        timeout = 1, 
                        true_addr = QFAM_equates.WORK_MODBUS_ADDR, 
                        not_addr = 0, 
                        match_addr = 0, 
                        and_mask = 0x00000000, 
                        or_mask = 0xFFFFFFFF)       # Set modbus address to 1
                    
        except Exception as e:
            if debug_on:
                print('Exception while sending masks: port = {}, {}'.format(self.modbus_client.port, e))
            self.port_OK = False
            self.meter.comm_OK = False
            return (None)
            
        ret_code = True
    
        if pulse_type not in ['WH', 'VARH', 'VAH']:
            return (False)
        pt_val = {
                    'WH'    : 0, 
                    'VARH'  : 1, 
                    'VAH'   : 2
                } [pulse_type]
                
        reg_vals = [pulse_weight] + 6 * [pt_val]
        reg_addr = 1278
        for reg_val in reg_vals:
            try:
                if debug_on:
                    print('Sending {} to register {}'.format(reg_val,  reg_addr))
                result = self.modbus_client.write_register(reg_addr, reg_val,  unit = QFAM_equates.WORK_MODBUS_ADDR)
                if debug_on:
                    print('Response : {}'.format(result))
                
            except Exception as e:
                if debug_on:
                    print('Exception while writing registers: port = {}, {}'.format(self.modbus_client.port, e))
                result = None
                
            
            if result:
                if result.isError():
                    if debug_on:
                        print('result.isError:{}'.format(result))
                    result = None
                    
            if debug_on:
                print('result: {}'.format(result))
                
            if result == None:
                ret_code = False
            reg_addr += 1
        
        return (ret_code)
