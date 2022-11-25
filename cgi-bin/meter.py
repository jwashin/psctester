#VERSION = 2.01
#VERSION = 2.02      #added tmx4 support
#VERSION = 2.03      # does test for TMX4
#VERSION = 2.04      # add support for reading of phase diagnostics
#VERSION = 2.05      #  Add retries to KWH reading
#VERSION = 2.06      #  Fix switching of TMX5 to 19200
#VERSION = 2.07      # Support tmx3 and tmx1
#VERSION = 2.08      # Bug fix - allow longer floats from tmx5
#VERSION = 2.09      # Add x -d to log from ST4
#VERSION = 2.10      # allow ST4 to search for comm parameters
#VERSION = 2.11      # Improve ST4 interface
#VERSION = 2.12      # Fix problems with TMX3 interface
#VERSION = 2.13      # begin Support TMX5n family
#VERSION = 2.14      # begin support for TMX4 with modem
#VERSION = 2.15      # improve support for TMX5n
#VERSION = 2.16      # improve support for TMX5n
#VERSION = 2.17      # add detail lines to csv file
#VERSION = 2.18      # add support for s10t
#VERSION = 2.19      # make version check work for tmx5b (281xxxxx)
#VERSION = 2.20      # fix bugs in version check
#VERSION = 2.21      # Add phase diagnostics to logs for tmx4 and tmx5
#VERSION = 2.22      # Fix problem of inappropriate reading of channel 1 in MC3
#VERSION = 2.23      # Fix error in reading of channel 1 in MC3
VERSION = 2.24      # Improve logging for comm; Speed up login; Improve error handling for bad version

SNAME = 'meter.py'

ver_file = open('./script_ver_work', 'a')
ver_file.write('{} {}\n'.format(SNAME, VERSION))
ver_file.close()

import serial
import time
import datetime
import re
import logging
from hardware_interface import Tester_hardware_obj
from conf import ST4_serno_list
from logwriter import logwriter as lw
from caveman_funcs import range_lookup

ST4_serno_tried = [False for i in ST4_serno_list]
   
try:
    st4_file = open('last_st4_serno.txt', 'r')
    ST4_serno = (int(st4_file.read()))
    st4_file.close()
except Exception:
    ST4_serno = (ST4_serno_list[0])
        
    
def get_next_st4_serno():
    global ST4_serno
    ret_code = True
    for i in range(len(ST4_serno_list)):
        if ST4_serno_list[i] == ST4_serno:
            ST4_serno_tried[i] = True
            try:
                if (ST4_serno_tried[i+1]):
                    ret_code = False
                ST4_serno = ST4_serno_list[i+1]
            except Exception:
                if (ST4_serno_tried[0]):
                    ret_code = False
                ST4_serno = ST4_serno_list[0]
            break
    else:
        ST4_serno = ST4_serno_list[0]
     
    st4_file = open('last_st4_serno.txt', 'w')
    st4_file.write('{}'.format(ST4_serno))
    st4_file.close()
    return (ret_code)


Q_logger = logging.getLogger('QLC_acc_tstr')


class MeterError(Exception):
    pass


class notImplemented(Exception):
    pass


def ret_true():
    return True

class Meter_obj():


    def __init__(self,ttydev,
                 m_type='TMX5',
                 m_serno=0,
                 m_prot='unknown',
                 m_numph=0
                 ):
        self.ttydev = ttydev
        Q_logger.info('open serial port: {}'.format(ttydev))
        self.com = serial.Serial(self.ttydev,
                                baudrate=
                                    {
                                    'TMX1': 2400, 
                                    'TMX3': 2400, 
                                    'TMX4' : 2400,
                                    'TMX5' : 19200,
                                    'TMX5n' : 19200, 
                                    'MC5n' : 19200
                                    }[m_type], 
                                bytesize =8,
                                parity='N',
                                stopbits=1,
                                interCharTimeout = 0.1,
                                timeout = 0.5)
        self.r_buf_maxlen = 20000
        self.KWH_reading = []
        self.passed_all_tests = False
        self.meter_type = m_type
        self.meter_serno = m_serno
        self.meter_protocol = m_prot
        self.num_phases = m_numph
        self.child_data = []

        if not self.setup_meter():
            raise MeterError('Could not contact meter')
        return

    def Keep_Alive(self):
        def Keep_Alive_serial():
            self.send_cmd('md -{}p\r'.format('G' if self.meter_type == 'TMX4' else ''),  com_retries = 1)
            
        def Keep_Alive_PLC():
            self.send_cmd('x -C6 -x\r', com_retries = 1 )
            
        def Keep_Alive_none():
            time.sleep(2)
            
        {
        'TMX1' : Keep_Alive_PLC, 
        'TMX3' : Keep_Alive_PLC, 
        'TMX4' : Keep_Alive_serial,
        'TMX5' : Keep_Alive_none, 
        'TMX5n' : Keep_Alive_serial, 
        'MC5n' : Keep_Alive_serial
        }[self.meter_type]()

    def wait_seconds(self, wait_secs, wait_message="Waiting",  keep_alive=True):
        time_loop_ctr = 0
        end_time = datetime.datetime.now() + datetime.timedelta(seconds=wait_secs)
        while datetime.datetime.now() < end_time:
            secs_remaining = end_time - datetime.datetime.now()
            lw.volatilemessage('{}; Seconds remaining: {:.0f}'.format(wait_message, secs_remaining.total_seconds()))
            time_loop_ctr += 1
            if time_loop_ctr < 5:
                time.sleep(2)
            else:
                time_loop_ctr = 0
                if keep_alive:
                    self.Keep_Alive()
                else:
                    time.sleep(2)
                
        lw.volatilemessage('{}; Done'.format(wait_message))
        

    def setup_meter(self):

        def read_version():
            self.meter_protocol = 'invalid'
            if self.send_cmd('ver\r'):
                r = re.search(r'version|software',
                              self.r_buf.decode('utf-8'))
                if r:
                    if r.group() == 'software':
                        r = re.search(r'\r\n.{0,7} {1,8}([0-9a-fA-F]{8})',
                                      self.r_buf.decode('utf-8'))
                    else:
                        r = re.search(r'\r\n([0-9a-fA-F]{8})',
                                      self.r_buf.decode('utf-8'))
                    if r:
                        self.meter_version = r.group(1)
                        self.meter_protocol = 'unknown'
            return
                              

##                if self.send_cmd('ver\r'):
##                    self.meter_protocol = 'unknown'
##
##                    try:
###                        self.meter_version = re.search(r'([0-9a-fA-F]{8})',
###                                                   self.r_buf.decode('utf-8')).group(1)
##                        self.meter_version = re.search(r'.{0,7} {1,8}([0-9a-fA-F]{8})',
##                                                   self.r_buf.decode('utf-8')).group(1)
##                    except Exception:
##                        self.meter_protocol = 'invalid'
            






            
        ret_code = False
        if self.login():
            if self.meter_type == 'TMX5n':
                if self.send_cmd('ver\r'):
                    self.meter_protocol = 'unknown'

                    try:
                        self.meter_version = re.search(r'.{0,7} {1,8}([0-9a-fA-F]{8})',
                                                   self.r_buf.decode('utf-8')).group(1)
                    except Exception:
                        self.meter_protocol = 'invalid'

                    if self.meter_protocol == 'invalid':
                        self.meter_protocol = 'unknown'
                    elif self.meter_version[0] == '5':
                        self.meter_protocol = 'TMX5n'
                        if self.meter_version[1] == '1':
                            self.num_phases = 2
                        elif self.meter_version[1] == '5':
                            self.num_phases = {
                                               '1':1,
                                               '2':2,
                                               '3':3
                                               }[self.meter_version[3:4]]
                        elif self.meter_version[1] == '9':
                            self.num_phases = 24
                            self.meter_protocol = 'MC5n'
                            self.child_data += [cd[0:3] for cd in self._get_MC5n_child_data()]
                            
                        else:
                            self.meter_protocol = 'unknown'
                            
                    elif self.meter_version[0] == '6':
                        self.meter_protocol = 'TMX5n'
                        self.num_phases = {
                                           '1':1,
                                           '2':2,
                                           '3':3
                                           }[self.meter_version[3:4]]
                    
                            
                    if self.meter_protocol != 'unknown':
                        self.use_current_gating = True
                        ret_code = True

            elif self.meter_type == 'TMX5':
##                if self.send_cmd('ver\r'):
##                    self.meter_protocol = 'unknown'
##
##                    try:
###                        self.meter_version = re.search(r'([0-9a-fA-F]{8})',
###                                                   self.r_buf.decode('utf-8')).group(1)
##                        self.meter_version = re.search(r'.{0,7} {1,8}([0-9a-fA-F]{8})',
##                                                   self.r_buf.decode('utf-8')).group(1)
##                    except Exception:
##                        self.meter_protocol = 'invalid'

                    read_version()

                    if self.meter_protocol == 'invalid':
                        self.meter_protocol = 'unknown'
                    else:
                        if self.meter_version[0] == '2':
                            self.meter_protocol = 'TMX5B'
                        elif self.meter_version[0] == '3':
                            if self.meter_version[4:8] <= '0110':
                                self.meter_protocol = 'TMX5C'
                            else:
                                self.meter_protocol = 'TMX5D'

                        if self.meter_protocol != 'unknown':

                            self.num_phases = {
                                               'a':24,
                                               '1':1,
                                               '2':2,
                                               '3':3
                                               }[self.meter_version[3:4]]
                            self.use_current_gating = False
                            ret_code = True
     
##          NOTE:  Folowing code is temporary, to allow 'TMX5 test' to work with tmx5n


                        if not ret_code:
                            ret_code = True
                            
                            if self.meter_version[0] == '5':
                                self.meter_protocol = 'TMX5n'
                                self.meter_type = 'TMX5n'
                                if self.meter_version[1] == '1':
                                    self.num_phases = 2
                                elif self.meter_version[1] == '5':
                                    self.num_phases = {
                                                       '1':1,
                                                       '2':2,
                                                       '3':3
                                                       }[self.meter_version[3:4]]
                                elif self.meter_version[1] == '9':
                                    self.num_phases = 24
                                    self.meter_protocol = 'MC5n'
                                    self.child_data += [cd[0:3] for cd in self._get_MC5n_child_data()]
                                    
                                else:
                                    self.meter_protocol = 'unknown'
                                    ret_code = False
                                    raise MeterError('Bad Meter Version: {}'.format(self.meter_version))
                                    
                            elif self.meter_version[0] == '6':
                                self.meter_protocol = 'TMX5n'
                                self.meter_type = 'TMX5n'
                                self.num_phases = {
                                                   '1':1,
                                                   '2':2,
                                                   '3':3
                                                   }[self.meter_version[3:4]]
                            else:
                                self.meter_protocol = 'unknown'
                                ret_code = False
                                
                            
                                
                        if (self.meter_protocol != 'unknown') and (self.meter_type == 'TMX5n'):
                            self.use_current_gating = True

                    

            elif self.meter_type == 'TMX4':
                if self.send_cmd('ver\r'):
                    reply_lines = self.r_buf.decode('utf-8').split('\r\n')
                    self.meter_version = reply_lines[1][10:18]

                    if self.send_cmd('md -Gp\r'):
                        reply_lines = self.r_buf.decode('utf-8').split('\r\n')
                        self.num_phases = 0
                        for line in reply_lines:
                            if re.search(r'(\s{1,2}\d{1,2}){2}',line):
                                self.num_phases += 1
                        if self.num_phases > 0:
                            if self.num_phases < 4:
                                self.meter_protocol = 'TMX4'
                            else:
                                self.meter_protocol = 'TMX4M'
                            self.use_current_gating = True
                            ret_code = True

            elif self.meter_type == 'TMX3':

                self.use_current_gating = True
                ret_code = True

            elif self.meter_type == 'TMX1':

                self.use_current_gating = True
                ret_code = True

        return ret_code

    def _wait_ST4_complete(self,time_limit=30,check_all=False):
        ret_code = False
        not_done = True
        end_time = time.time()+time_limit
        ST4_wait_counter = 1
        lw.message('Try PLC comm {}'.format(ST4_wait_counter))
        while not_done:

#            self.send_cmd('dt\r',cmd_timeout=10)
            self.send_cmd('x -d\r',com_retries = 1)
            lw.changelast('Try PLC comm {}'.format(ST4_wait_counter))
            ST4_wait_counter += 1

            if self.send_cmd('xr -Gd\r',cmd_timeout=10):

                reply_lines = self.r_buf.decode('utf-8').split('\r\n')
                if  ((len(reply_lines) < 4) or
                     reply_lines[1][:12] != 'Mtr location'):
                    raise MeterError('bad reply reading xref')
                else:
                    for work_line in reply_lines[2:(-1 if check_all else 3)]:
#                        print(work_line)
                        ST4_xref_stat = re.search(
                            r'(\b[0-9a-fA-F]{4}\b) (\b[0-9a-fA-F]{4}\b) (\b[0-9a-fA-F]{4}\b) (\b[0-9a-fA-F]{4}\b) ',
                            work_line
                            )
                        if not ST4_xref_stat:
                            raise MeterError('bad data in xref line')
                        else:
                            if ST4_xref_stat.group(1) != '0000':
                                break
                    else:
                        ret_code = True
                        not_done = False

            else:
                not_done = False

            if time.time() > end_time:
                not_done = False

        return(ret_code)

    def _try_login_ST4(self, pre_check=False):
        if (pre_check and self.send_cmd('attn -d\r',com_retries=1)):
            ret_code = True
        else:
            ret_code = False
            try:
                if self.send_cmd('attn -S{} -4u5574\r'.format(ST4_serno)):
                    ret_code = True
            except:
                pass

        return(ret_code)

    def _setup_ST4(self):
        ret_code = False
        try:
            ret_code = (self.send_cmd('scan 0\r',cmd_timeout=10,com_retries=1) and
                        self.send_cmd('xr -GZ1234\r',cmd_timeout=10,com_retries=1) and
                        self.send_cmd('sm 1\r') and
                        self.send_cmd('allow 3107\r') and
                        self.send_cmd('xr -A3107 -i3107\r') and
                        self.send_cmd('da -l1\r') and
                        self.send_cmd('scan\r')
                        )
        except:
            pass

        return(ret_code)
        
    def _get_MC5n_child_data(self):
        """ reads values from child meters on an MC5n.  Returns list of tuples.  
            One tuple for each channel. Each tuple contains 4 string values:
            (Slot, Child Serial Number, Child Phase, and KWH)
        """

        retries = 3
        while (retries > 0):
            
            work_child_data = []
            
            try:
                self.send_cmd('mscan -Gt\r')
                reply_lines = self.r_buf.decode('utf-8').split('\r\n')
                if ((len(reply_lines) != self.num_phases + 3) or 
                    (reply_lines[1][:10] != '$lot serno')):
                        
                    raise MeterError('bad reply reading KWH')
                else:
                    for work_line in reply_lines[2:-1]:
                        r = re.search(r' (  \d| \d\d|\d\d\d) ([ \d]{8})[ \d]{6} ([ \d]{3})  {0,5}(\-?[\d\.]{11,16})',work_line)
                        if r:
                            work_child_data += [r.group(1, 2, 3, 4)]
                        else:
                            raise MeterError('Bad data line in KWH readings')
                    retries = 0
            except MeterError:
                retries -= 1
                if (retries <= 0):
                    raise

        return work_child_data

#                    r = re.search(r'[ \d]{3}\d ([ \d]{8})[ \d]{6} ([ \d]{3})  {0,5}(\-?[\d\.]{11,16})',work_line)

    def login(self):

        def setup_TMX1_and_3():
            self.channel_mask = [True]
            try:
                ret_code = self.setup_PLC_comm()
            except:
                ret_code = False
                pass

            if ret_code:
                self.meter_version = re.search(r"\b([0-9a-fA-F]{8}) 0000 ",
                                               self.r_buf.decode('utf-8')
                                               ).group(1)
                ret_code = self.meter_version != "00000000"

            return(ret_code)

        def do_login_TMX5():

            def try_login():
                ret_code = False
                attn_D_re = r'(\d{8})( {1,4}[0-9A-Fa-f]{1,4}){1,2}\r?'
                if self.send_cmd('attn -D\r',
                                 cmd_timeout = 2, 
                                 com_retries=4,
                                 prompt_re=attn_D_re):
                    self.meter_serno = eval(re.search(attn_D_re,self.r_buf.decode('utf-8')).group(1))
                    try:
                        login_re = r' at {1,2}(\d{1,2}\:\d\d\:\d\d) {1,2}(\d{1,2}/\d\d/\d\d\d\d)'
                        if  (self.send_cmd('attn -S{} -5lEvElbAl\r'.format(self.meter_serno), prompt_re=login_re) or
                             self.send_cmd('attn -S{} -74321fdsa\r'.format(self.meter_serno), prompt_re=login_re)):
                            (self.meter_time,self.meter_date) = re.search(
                                                                    login_re,
                                                                    self.r_buf.decode('utf-8')).group(1,2)
                            ret_code = True
                    except Exception:
                        ret_code = False

                return(ret_code)

            ret_code = False
            self.set_baudrate(19200)
            if try_login():
                ret_code = True
            else:
                self.set_baudrate(9600)
                if try_login():
                    if self.send_cmd('stty 19200\r', cmd_flags='n'):
                        time.sleep(0.2)
                        self.set_baudrate(19200)
                        ret_code = self.send_cmd('\r', com_retries=3)

            return(ret_code)

        def do_login_TMX4():

            def try_login():
                ret_code = False
                try:
                    if self.send_cmd('attn -S{} -4u5574\r'.format(self.meter_serno)):
                        (self.meter_time,self.meter_date) = re.search(
                            r' at {1,2}(\d{1,2}\:\d\d\:\d\d) {1,2}(\d{1,2}/\d\d/\d\d\d\d)',
                            self.r_buf.decode('utf-8')).group(1,2)
                        ret_code = True
                except Exception:
                    pass

                return(ret_code)

            ret_code = False

            self.set_baudrate(2400)
            if try_login():
                if self.send_cmd('baudhigh\r', cmd_flags='n'):
                    time.sleep(0.2)
                    self.set_baudrate(9600)
                    ret_code = self.send_cmd('\r',com_retries=3)

            else:
                self.set_baudrate(9600)
                if try_login():
                    ret_code = True
                    
            return(ret_code)

        def do_login_TMX1_and_3():
            login_success = False
            while (not login_success):
                self.set_baudrate(2400)
                if self._try_login_ST4():
                    if self.send_cmd('baudhigh\r'):
                        time.sleep(0.2)
                        self.set_baudrate(9600)
                        login_success = self.send_cmd('\r',com_retries=3)
                        
                if not login_success:
                    self.set_baudrate(9600)
                    if self._try_login_ST4():
                        login_success = True
                        
                if (not login_success):
                    msg_string = 'Failed login to ST4# {} - '.format(ST4_serno)
                    if (get_next_st4_serno()):
                        lw.message('{} trying {}'.format(msg_string, ST4_serno))
                    else:
                        lw.message('{} all ST4s failed'.format(msg_string))
                        break

            if login_success:
                login_success = setup_TMX1_and_3()

            return(login_success)

#   Main login() code begins here

        if  (self.meter_type in ('TMX4', 'TMX5') and
             self.send_cmd('attn -d\r',com_retries=1)):
            ret_code = True
        else:
            ret_code = {
                        'TMX1': do_login_TMX1_and_3,
                        'TMX3': do_login_TMX1_and_3,
                        'TMX4': do_login_TMX4,
                        'TMX5': do_login_TMX5, 
                        'TMX5n': do_login_TMX5
                        } [self.meter_type]()
        return(ret_code)

    def set_baudrate(self,baud_rate):
        Q_logger.info('set baud rate: {}'.format(baud_rate))
        self.com.baudrate = baud_rate
        return()

    def send_cmd(self,
                 cmd_str,
                 cmd_flags='r',
                 cmd_timeout=3,
                 com_retries=3,
                 prompt_re=r'(CIP[\:>#\$\\]\r?$)|(S[\:>#\$\\]\r?$)',
                 data_validator=ret_true
                 ):
        #   cmd_timeout is time in seconds for line quiet (no chars rcvd).  Will not return
        #   until quiet time met, or prompt found, or buffer overflow.  May wait forever.

        def restart_timeout():
            return (time.time() + cmd_timeout)

        short_wait_time = 0.2
        ret_code = False
        s_buf = cmd_str.encode('utf-8')
        modem_retries = 5

        while com_retries:

            com_retries -= 1
            self.com.flushInput()
            self.r_buf = b''
#            print(s_buf)
            self.com.write(s_buf)
            Q_logger.info('send: {}'.format(s_buf.decode('utf-8'))) ###
            if cmd_flags == 'n':
                # no reply expected - return OK
                ret_code = True
                break
            else:
                # reply expected - receive to buffer
                self.com.timeout = short_wait_time
#                self.com.interCharTimeout = short_wait_time
#                print('get reply')
                timeout_limit = restart_timeout()

                while ((time.time() < timeout_limit) and
                       (len(self.r_buf) < self.r_buf_maxlen)):

#                    print('read')
                    rd_char = self.com.read(self.r_buf_maxlen - len(self.r_buf))
#                    print(len(rd_char),type(rd_char),rd_char)

                    try:
                        rd_char.decode('utf-8')     # Check for invalid utf-8 codes
                    except:
#                        print('except:',len(rd_char),type(rd_char),rd_char)
                        rd_char = b''

#                    print(len(rd_char),rd_char)

                    if len(rd_char) > 0:
                        # got some characters - reset timeout, add to buffer, and see if prompt
                        timeout_limit = restart_timeout()
                        self.r_buf += rd_char
#                        print(len(self.r_buf),self.r_buf)
                        if re.search(prompt_re, self.r_buf.decode('utf-8')):

                            if not re.search(r'[Ee]rror', self.r_buf.decode('utf-8')):
                                if data_validator():
                                    ret_code = True
                                    com_retries = 0

                            break
                            
                        elif re.search('modem failure', self.r_buf.decode('utf-8')):
                            time.sleep(25)
                            if cmd_str != 'attn -d\r':
                                com_retries += modem_retries
                                modem_retries = 0

            Q_logger.info('receive: {}\n{}'.format(ret_code,self.r_buf.decode('utf-8')))  ###
        return ret_code

    def get_KWH(self):

        def get_KWH_TMX5n():
            """ reads KWH accumulator values from a TMX5n single-point meter.
            returns a list of KWH values with one entry per meter
            phase"""

            work_KWH_reading = []

            self.send_cmd('md -t\r')
            reply_lines = self.r_buf.decode('utf-8').split('\r\n')
            if ((len(reply_lines) != (self.num_phases * (self.num_phases + 1) + 3))
                or not re.match(r' p# {7,12}kWH',reply_lines[1])):
                raise MeterError('bad reply reading KWH')
            else:
                for work_line in reply_lines[2:-2:self.num_phases + 2]:
                    r = re.search(r'[ \d]\d {0,4}(\-?[\d\.]{11,16})',work_line)
                    if r:
                        work_KWH_reading += [eval(r.group(1))]
                    else:
                        raise MeterError('Bad data line in KWH readings')

            return work_KWH_reading

        def get_KWH_MC5n():
            """ reads KWH accumulator values from a TMX5n minicloset.
            returns a list of KWH values with one entry per meter
            phase"""

#            work_KWH_reading = []
#
#            self.send_cmd('mscan -Gt\r')
#            reply_lines = self.r_buf.decode('utf-8').split('\r\n')
#            if ((len(reply_lines) != self.num_phases + 4) or 
#                (reply_lines[1][:10] != '$lot serno')):
#                    
#                raise MeterError('bad reply reading KWH')
#            else:
#                for work_line in reply_lines[2:-2]:
##                    r = re.search(r'[ \d]{3}\d ([ \d]{8})[ \d]{6} ([ \d]{3})  {0,5}(\-?[\d\.]{11,16})',work_line)
#                    r = re.search(r'[ (  \d|  \d\d|\d\d\d) ([ \d]{8})[ \d]{6} ([ \d]{3})  {0,5}(\-?[\d\.]{11,16})',work_line)
#                    if r:
#                        work_KWH_reading += [eval(r.group(4))]
#                    else:
#                        raise MeterError('Bad data line in KWH readings')
#

#            work_child_data = self._get_MC5n_child_data()
#            work_KWH_reading = [eval(cd[3]) for cd in work_child_data]
            work_KWH_reading = [eval(cd[3]) for cd in self._get_MC5n_child_data()]
            return work_KWH_reading

        def get_KWH_TMX5D():
            """ reads KWH accumulator values from a TMX5 meter.
            returns a list of KWH values with one entry per meter
            phase"""

            work_KWH_reading = []

            self.send_cmd('md -tc\r')
            reply_lines = self.r_buf.decode('utf-8').split('\r\n')
            if len(reply_lines) != self.num_phases + 3 or not re.match(r'(Mtr PH)|( M# p# )', reply_lines[1]):
                raise MeterError('bad reply reading KWH')
            else:
                for work_line in reply_lines[2:-1]:
                    r = re.search(r'(?: [ \d]\w){3}( {5}[ \-][\d\.]{11}){4}',work_line)
                    if r:
                        work_KWH_reading += [eval(r.group(1))]
                    else:
                        raise MeterError('Bad data line in KWH readings')

            return work_KWH_reading

        def get_KWH_TMX5B():
            """ reads KWH accumulator values from a TMX5 meter.
            returns a list of KWH values with one entry per meter
            phase"""

            work_KWH_reading = []

            self.send_cmd('md -Gt -Q2\r')
            reply_lines = self.r_buf.decode('utf-8').split('\r\n')
#            print(self.r_buf)
            if len(reply_lines) != self.num_phases + 3 or reply_lines[1][:6] != 'Mtr PH':
                raise MeterError('bad reply reading KWH')
            else:
                for work_line in reply_lines[2:-1]:
#                    print(work_line)
                    hold_result = re.search(r'( [ \d]\d){3}( {1,8}\-?[\d\.]{1,17})' +
                                            r'( {1,8}\-?[\d\.]{1,17})' +
                                            r'( {1,8}\-?[\d\.]{1,17})',
                                            work_line)
                    if hold_result:
#                        print(hold_result.groups())
                        hold_KWH = eval(hold_result.group(2)) + \
                                   eval(hold_result.group(3)) + \
                                   eval(hold_result.group(4))
#                        print(hold_KWH)
                        work_KWH_reading += [hold_KWH]
                    else:
                        raise MeterError('Bad data line in KWH readings')

            return work_KWH_reading

        def get_KWH_TMX4():
            """ reads KWH accumulator values from a TMX4 meter.
            returns a list of KWH values with one entry per meter
            phase"""

            cmd_retries = 3

            while cmd_retries > 0:

                work_KWH_reading = []       # Clear result list

                try:

                    self.send_cmd('md -Gt -Q2\r')
                    reply_lines = self.r_buf.decode('utf-8').split('\r\n')
        #            print(self.r_buf)
                    if len(reply_lines) != self.num_phases + 3 or reply_lines[1][:6] != 'Mtr PH':
                        raise MeterError('bad reply reading KWH')
                    else:
                        for work_line in reply_lines[2:-1]:
        #                    print(work_line)
                            hold_result = re.search(r'( [ \d]\d){2}( {1,8}\-?[\d\.]{1,13})' +
                                                    r'( {1,8}\-?[\d\.]{1,13})',
                                                    work_line)
                            if hold_result:
        #                        print(hold_result.groups())
                                hold_KWH = eval(hold_result.group(2))
        #                        print(hold_KWH)
                                work_KWH_reading += [hold_KWH]
                            else:
                                raise MeterError('Bad data line in KWH readings')

                        cmd_retries = 0

                except:
                    cmd_retries -= 1
                    if cmd_retries == 0:
                        raise


            return work_KWH_reading

        def get_KWH_TMX3M():
            """ Reads KWH accumulator values from an MC-3 meter, using an ST4.
            Returns a list of KWH values with one entry per meter
            phase.
            Only phases which are True in the channel_mask will have valid values.
            All other phases will report 0.0.
            """

            def read_KWH_TMX3M():

                got_readings = False
                retry_limit = 3

                while not got_readings:

                    try:

                        if not self._try_login_ST4(pre_check=True):
                            raise MeterError('Could not communicate with ST4')

                        work_KWH_reading = self.num_phases * [0.0]
                        work_active_phase_list = [i+1 for i in range(len(self.channel_mask)) if (i == 0 or
                                                                                                 self.channel_mask[i])]
##                        work_active_phase_list = [1]
##            #            print(*self.channel_mask,sep='\n')
##                        for i in range(1,len(self.channel_mask)):
##                            if self.channel_mask[i]:
##                                work_active_phase_list += [i+1]

                        work_active_phases = len(work_active_phase_list)

                        if  (self.send_cmd('xr -G -R1000\r') and
                             self._wait_ST4_complete(time_limit=7*work_active_phases+20,check_all=True)
                             ):
                            try_md_GT = True
                            md_GT_retries = 3
                            while try_md_GT:
                                found_bad_data = True
                                if self.send_cmd('md -GT\r',cmd_timeout=10):
                                    reply_lines = self.r_buf.decode('utf-8').split('\r\n')
                #                    print(*reply_lines,sep='\n')

                                    if  (len(reply_lines) == work_active_phases + 3 and
                                         reply_lines[1][:12] == 'Mtr location'):
                                        found_bad_data = False
                                        for work_line_index in range(work_active_phases):
                                            work_line = reply_lines[work_line_index+2]
                        #                    print(work_line)
                                            hold_result = re.search(r'( [ \d]\d)\s{10}(\d{8})( {1,10}\-?[\d\.]{1,10}) kWH',
                                                                    work_line)
                                            if hold_result:
                #                                print(hold_result.groups())
                                                hold_KWH = eval(hold_result.group(3))
                                                hold_index = eval(hold_result.group(2)) - self.meter_serno
                #                                print(hold_index)
                                                if  (hold_index == work_active_phase_list[work_line_index] - 1):
                                                    
                                                    if self.channel_mask[hold_index]:
                                                        work_KWH_reading[hold_index] = hold_KWH
                                                else:
                                                    found_bad_data = True
                                                    break
                                            else:
                                                found_bad_data = True
                                                break

                                if not found_bad_data:
                                    try_md_GT = False
                                    got_readings = True
                                else:
                                    md_GT_retries -= 1
                                    try_md_GT = md_GT_retries > 0

                            if found_bad_data:
                                raise MeterError('Unable to read KWH from ST4')

                        else:
                            raise MeterError('Unable to read KWH via PLC')

                    except MeterError:
                        retry_limit -= 1
                        if retry_limit == 0:
                            raise


    #            print(*work_KWH_reading,sep='\n')
                return work_KWH_reading

            #
            #
            ################### Main code for get_KWH_TMX3M   ################
            #
            #

            first_val = read_KWH_TMX3M()
            second_val = read_KWH_TMX3M()
            if first_val != second_val:
                third_val = read_KWH_TMX3M()
                if first_val != third_val:
                    if second_val == third_val:
                        first_val = third_val
                    else:
                        raise MeterError('Inconsistent readings from MC3 via PLC')

            return first_val


        def get_KWH_TMX3_TMX1():
            """ Reads KWH accumulator values from a TMX3 or TMX1 meter, using an ST4.
            Returns a list of KWH values with one entry per meter
            phase"""
            
#
            def get_KWH_total():
                
                hold_KWH=0.0
                    
                def chk_st4_KWH_data():
                    nonlocal hold_KWH
                    ret_code = False
                    try:
                        hold_KWH = eval(re.search(r'^\s*(-?\d*\.?\d*) kWH',
                                            self.r_buf.decode('utf-8').split('\r\n')[2]).group(1))
                        ret_code = True
                    except Exception:
                        pass
                    return (ret_code)
                    
                if (self.send_cmd('xr -M1 -R1000\r') and
                    self._wait_ST4_complete()
                    ):

                    if not self.send_cmd('md -T\r',
                                        cmd_timeout = 4, 
                                        data_validator = chk_st4_KWH_data, 
                                        com_retries=5):
                    
                        raise MeterError('Error getting KWH total from ST4')

                    return(hold_KWH)

                else:
                    raise MeterError('Problem with ST4')
                    
            if not self._try_login_ST4(pre_check = True):
                raise MeterError('Could not communicate with ST4')

            first_val = get_KWH_total()
            second_val = get_KWH_total()
            if first_val != second_val:
                third_val = get_KWH_total()
                if first_val != third_val:
                    if second_val == third_val:
                        first_val = third_val
                    else:
                        raise MeterError('Inconsistent readings from meter via PLC')

            return [first_val/3, first_val/3, first_val/3]


#
#       Main code for getKWH()
#

        work_KWH = {
                'TMX5B' : get_KWH_TMX5B,
                'TMX5C' : get_KWH_TMX5B,
                'TMX5D' : get_KWH_TMX5D,
                'TMX4'  : get_KWH_TMX4,
                'TMX4M' : get_KWH_TMX4,
                'TMX1' : get_KWH_TMX3_TMX1,      ## test !!!!!!
                'TMX3' : get_KWH_TMX3_TMX1,
                'TMX3M' : get_KWH_TMX3M, 
                'TMX5n' : get_KWH_TMX5n, 
                'MC5n'  : get_KWH_MC5n
                        }[self.meter_protocol]()
        Q_logger.debug('KWHv: ' + ' '.join(map(str,work_KWH)) + '\n')
        return work_KWH

    def prepare_for_KWH(self):

        def prepare_for_KWH_TMX5D():
            return True

        return {
                'TMX5B' : prepare_for_KWH_TMX5D,
                'TMX5C' : prepare_for_KWH_TMX5D,
                'TMX5D' : prepare_for_KWH_TMX5D,
                'TMX4' : prepare_for_KWH_TMX5D,
                'TMX4M' : prepare_for_KWH_TMX5D,
                'TMX3' : prepare_for_KWH_TMX5D,
                'TMX3M' : prepare_for_KWH_TMX5D,
                'TMX1'  : prepare_for_KWH_TMX5D, 
                'TMX5n' : prepare_for_KWH_TMX5D,
                'MC5n' : prepare_for_KWH_TMX5D,
                        }[self.meter_protocol]()

    def freeze_reg(self):

        def freeze_reg_TMX5D():

            self.send_cmd('mt -g-1 -H\r')

            return self.r_buf.decode('utf-8')[:27] == '\r\nI Off - Wait new cmd\r\nCIP'

        return {
                'TMX5B' : freeze_reg_TMX5D,
                'TMX5C' : freeze_reg_TMX5D,
                'TMX5D' : freeze_reg_TMX5D
                        }[self.meter_protocol]()


    def arm_reg_gate(self):

        def arm_reg_gate_TMX5D():

            return self.send_cmd('mt -g-1\r',prompt_re='I Off - Wait 1st char')

        return {
                'TMX5B' : arm_reg_gate_TMX5D,
                'TMX5C' : arm_reg_gate_TMX5D,
                'TMX5D' : arm_reg_gate_TMX5D
                        }[self.meter_protocol]()

    def enable_reg(self):

        def enable_reg_TMX5D():

            return self.send_cmd('mt -H\r')

        return {
                'TMX5B' : enable_reg_TMX5D,
                'TMX5C' : enable_reg_TMX5D,
                'TMX5D' : enable_reg_TMX5D
                        }[self.meter_protocol]()

    def setup_PLC_comm(self):

        def setup_PLC_comm_TMX1():

            ret_code = False

            if (self._setup_ST4()):
                tmx1_software_code = range_lookup(self.meter_serno, 
                                                    [
                                                    [83070000,  84109999, 1],
                                                    [84110000,  84129999, 2],
                                                    [85010000,  87039999, 3],
                                                    [87040000,  88079999, 4],
                                                    [80200000,  80309999, 4],                                                    
                                                    ])
                ret_code = (
                            self.send_cmd('xr -M1 -R0007 -T1 -s{} -I2 -S{} -H1 -D3 -Y8\r'.format(self.meter_serno,  
                                                                                                tmx1_software_code)) and
                            self._wait_ST4_complete(time_limit=300)
                            )


            return ret_code

        def setup_PLC_comm_TMX3():

            return  (self._setup_ST4() and
                    self.send_cmd('xr -M1 -s{} -R0007 -T1 -I2 -S5 -H0 -D3 -Y8\r'.format(self.meter_serno)) and
                    self._wait_ST4_complete(time_limit=120))


        return {
                'TMX1' : setup_PLC_comm_TMX1,
                'TMX3' : setup_PLC_comm_TMX3,
                'TMX3M' : setup_PLC_comm_TMX3
                        }[self.meter_protocol]()


    def setup_active_channel_scan(self):

        work_meter = 0
        for i in range(len(self.channel_mask)):

            if self.channel_mask[i] or (i == 0):
                work_meter += 1
                if not (
                    self.send_cmd('sm {}\r'.format(work_meter)) and
                    self.send_cmd('xr -M{} -P1 -m{} -R0000 -A3107 -i3107\r'.format(work_meter, i+1))
                    ):
                    return False

        return (self.send_cmd('mmult -GI 1 1\r') and
                self.send_cmd('mmult -GV 1 1\r'))




### Read phase diagnostics  ###

    def get_phase_data(self):

        def get_phase_data_TMX4():
            """
            Reads KWH accumulator values from a TMX4 meter.
            Returns a list of lists of phase realtime data values.
            Returns a list [Amps,Volts,Watts,VARs] for each meter phase.
            """

            work_phase_reading = []

            self.send_cmd('md -Gp\r')
            reply_lines = self.r_buf.decode('utf-8').split('\r\n')
            if len(reply_lines) != self.num_phases + 3 or reply_lines[1][:6] != 'Mtr PH':
                raise MeterError('bad reply reading phase')
            else:
                for work_line in reply_lines[2:-1]:
                    hold_result = re.search(r' [ \d]\d [ \d]\d [0-9a-fA-F]' +
                                            r'( {0,8}\-?[\d\.]{1,13})([ muk])A' +
                                            r'( {0,8}\-?[\d\.]{1,13})([ muk])V' +
                                            r'( {0,8}\-?[\d\.]{1,13})([ muk])W' +
                                            r'( {0,8}\-?[\d\.]{1,13})([ muk])VAR',
                                            work_line)
                    if hold_result:
                        hold_phase = []
                        for i in range(4):
                            j = 2*i+1
                            hold_phase += [eval(hold_result.group(j)) *
                                           {' ' : 1.0,
                                            'm' : 0.001,
                                            'u' : 0.000001,
                                            'k' : 1000.0
                                            } [hold_result.group(j+1)]
                                           ]
                        work_phase_reading += [hold_phase]
                    else:
                        raise MeterError('Bad data line in phase readings')

            return work_phase_reading

        def get_phase_data_TMX3M():
            """
            Reads Phase Diagnostic values from a TMX3M meter using an ST4.
            Returns a list of lists of phase realtime data values.
            Returns a list [Amps,Volts,Watts,VARs] for each meter phase.
            Since the TMX3 does not report phase diagnostics, the returned lists
            contain the following: [estimatedAmps,120.0,estimatedWatts,0].
            estimatedAmps is calculated by reading the KWH twice, 30 seconds apart.
            These KWH readings are subtracted to get estimatedAmps.
            estimatedWatts = estimatedAmps * 120.
            """

            work_phase_reading = []

            start_readings = self.get_KWH()

            time.sleep(30)

            end_readings = self.get_KWH()

            for i in range(len(start_readings)):
                estimatedAmps = (end_readings[i] - start_readings[i]) * 200.0
                estimatedWatts = estimatedAmps * 120.0
                work_phase_reading += [[estimatedAmps,120.0,estimatedWatts,0.0]]

            return work_phase_reading

        return  {
                'TMX3M'  : get_phase_data_TMX3M,
                'TMX4'  : get_phase_data_TMX4,
                'TMX4M' : get_phase_data_TMX4
                        }[self.meter_protocol]()



if __name__ == '__main__':

    import operator


    print('start')

    tester = Tester_hardware_obj()

    tester.set_volts_off()
    time.sleep(1)
    tester.set_volts_on()
    tester.preset_amps_high()
    tester.set_amps_on()


    try:
        DUT = Meter_obj('/dev/tty_meter')
        Okay = True
    except MeterError:
        print ('Unable to contact meter')
        Okay = False

    if Okay:

        print(DUT.get_phase_data())
        print(DUT.prepare_for_KWH())
        print(DUT.freeze_reg())
        print(DUT.arm_reg_gate())
        time.sleep(5)
        DUT.send_cmd('a',prompt_re='2nd char',cmd_timeout=5)
        time.sleep(5)
        DUT.send_cmd('b')
        print(DUT.enable_reg())

        DUT.send_cmd('dt\n')

        start_reading = DUT.get_KWH()

#        print(start_reading)

        time.sleep(5)

        end_reading = DUT.get_KWH()

##        for w in range(len(start_reading)):
##            end_reading[w] -= start_reading[w]

        end_reading = map(operator.sub,end_reading,start_reading)

        print(list(end_reading))

        print(DUT.meter_type, DUT.meter_serno, DUT.meter_time, DUT.meter_date, DUT.meter_version, DUT.num_phases)


