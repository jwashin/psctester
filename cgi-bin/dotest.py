##VERSION = 2.0
##VERSION = 2.01  # Add empty cell after test date in report file
##VERSION = 2.02  # Correct test times.  2.01 had invalid times
##VERSION = 2.03      # update interface to new GUI
                    # support TMX4 testing
##VERSION = 2.04      # Support partial tests for MC-4
##VERSION = 2.05      # Removed spurious heading
##VERSION = 2.06      #  Add read of tester config
##VERSION = 2.07      # Use new time delay routine
##VERSION = 2.08      # Add exception trace to log file
##VERSION = 2.09      # Check for at least 2 channels on TMX4
##VERSION = 2.10      # Report meter errors rather than crashing the script
##VERSION = 2.11      # Support TMX3
##VERSION = 2.12      # support tmx1 (not completed)
##VERSION = 2.13      # support incremental tests
##VERSION = 2.14      # support tmx5n
##VERSION = 2.15      # improve tmx5n support
##VERSION = 2.16      # add detail lines to csv file
##VERSION = 2.17      # add applicable date and number of test points to csv file
VERSION = 2.18      # add support for built-in rs-485 port 



SNAME = "dotest.py"

ver_file=open('./script_ver_work','w')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

from subprocess import Popen, PIPE
import time
from logwriter import logwriter as lw
import json
import sys
import os

import datetime
import logging
import logging.handlers
from filesloc import filesloc
from conf import tester_number

from tests import *

from test_data_log import test_data_log

#
#   Test_times tuples = (low current seconds, high current seconds, meter update seconds)
#
##test_times = {"TMX5B" : (90,20,30),
##              "TMX5C" : (90,20,30),
##              "TMX5D" : (90,20,30),
##              "TMX4"  : (100,45,10),
##              "TMX4M" : (200,150,10),
##              "TMX3" : (100,300,15),
##              "TMX3M" : (3000,3000,30),
##              "TMX1" : (100,300,15)
##              }

def Passed(err_percent):
##    if err_percent < 0:
##        err_percent = 0 - err_percent
    return abs(err_percent) < 1.8

def dotest():
    
    """
    Get data and report results via a .json file.
    
    Main routine for PSC tester.  Executed by web server as cgi-bin routine.
    
    """

    def log_KWH_results(report_title):

        logfile.write("\n{}:\n".format(report_title))
        
        lw.message("Test of {} at {:.2f} amps".format(DUT.meter_serno,Radian.amps*10.0))
        
        logfile.write("Test Current (Amps):,, {:.2f}\n".format(Radian.amps*10.0))
        
        if len(DUT.child_data) == 0:
            logfile.write("Channel,Percent Error{},,Within tolerance?\n".format(' (avg)' if DUT.num_active_phases <= 3 else ''))

            for chan_index in range(len(DUT.err_percent)):
                if DUT.channel_mask[chan_index]:
                    work_status = DUT.pass_fail[chan_index]
                    work_percent_err = DUT.err_percent[chan_index]
                    lw.message("Channel {:2}: {:7.2f} percent error - {}{}".format(chan_index+1,
                                                                        DUT.raw_err_percent[chan_index],
                                                                        work_status, 
                                                                        ' (avg)' if DUT.num_active_phases <= 3 else ''))
                    logfile.write("{},{:6.2f},,{}\n".format(chan_index+1,
                                                            work_percent_err,
                                                            work_status))
                                                            
        else:
            logfile.write("Channel,Percent Error,,Within tolerance?,,,Child Serno,,Child channel\n")

            for chan_index in range(len(DUT.err_percent)):
                if DUT.channel_mask[chan_index]:
                    work_status = DUT.pass_fail[chan_index]
                    work_percent_err = DUT.err_percent[chan_index]
                    lw.message('Channel {:2} ({}{}): {:7.2f} percent error - {}'.format(chan_index+1,
                                                                                            DUT.child_data[chan_index][1],
                                                                                            DUT.child_data[chan_index][2],
                                                                                            work_percent_err,
                                                                                            work_status))
                    logfile.write("{},{:6.2f},,{},,,{},,{}\n".format(chan_index+1,
                                                                    work_percent_err,
                                                                    work_status, 
                                                                    DUT.child_data[chan_index][1],
                                                                    DUT.child_data[chan_index][2]))
                                                                            
        logfile.flush()
        return True

    def log_test_details():

        logfile.write('Test data details:\n')
        
        logfile.write("Amps,,Seconds,,actual KWH,,Start read,,End read,,Error %,,Uncertainty\n")
        
        for test_key in ['low', 'high']:
            for data_index in range (len(test_data_log[test_key].phase_data)):
                logfile.write('{},,{},,{},,{},,{},,{},,{}\n'.format(test_data_log[test_key].test_amps, 
                                                                    test_data_log[test_key].test_seconds, 
                                                                    test_data_log[test_key].radian_KWH, 
                                                                    test_data_log[test_key].phase_data[data_index][0], 
                                                                    test_data_log[test_key].phase_data[data_index][1], 
                                                                    test_data_log[test_key].phase_data[data_index][2], 
                                                                    test_data_log[test_key].percent_uncertainty
                                                                    ))
        
        logfile.flush()
        return True

    def check_MC_active_phases():
        ret_code = False
        off_phase_data = DUT.get_phase_data()
#        print(*off_phase_data)
        tester.preset_amps_high()
        tester.set_amps_on()
        DUT.wait_seconds(meter_update_delay)
        on_phase_data = DUT.get_phase_data()
        tester.set_amps_off()
#        print(*on_phase_data)
        num_found_channels = 0
        for i in range(DUT.num_phases):
            DUT.channel_mask[i] = (off_phase_data[i][0] < 0.5) and (on_phase_data[i][0] > 60)
#            print(i,DUT.channel_mask[i],off_phase_data[i],on_phase_data[i])
            if DUT.channel_mask[i]:
                lw.message('Tester current detected on channel {}'.format(i+1))
                num_found_channels += 1

#        print(*DUT.channel_mask)
        
        if num_found_channels < 2:
            lw.message('Only {} channels detected.  Check current connections.'.format(num_found_channels))
        else:
            ret_code = True
            
        return(ret_code)

    def dump_KWH():
        
        tester.preset_amps_high()
        tester.set_amps_on()
        
        for i in range(200):
            hold_KWH = DUT.get_KWH()[0]
            print(time.time(),hold_KWH)
            
        tester.set_amps_off()
        

#                                   #
#                                   #
#   Beginning of dotest main code   #
#                                   #
#                                   #


    thefile = '../control.json'

    if os.path.exists(thefile):
        try:
            with open(thefile,'r') as control:
                json.load(control)
        except:
            return False;
##        if data['status'] != 'done':
##            return False

    try:
        indata = json.loads(sys.stdin.read())
    except Exception:
        indata = {'serial': '10130',
                  'siteid':'92240',
                  'c':'tmx5',
                  'x3':'MC3-24',
                  'channels' : [22, 23, 24]
                  }           # test values when script started from idle

    site_ID = indata['siteid']

    serial = indata['serial']

    meter_protocol = 'unknown'
    meter_serno = 0
    meter_num_phases = 0        # Default values
    
    if indata['c'] == 'tmx1':
        test_type = 'TMX1'
        meter_serno = int(serial)   
        meter_num_phases = 3
        meter_protocol = 'TMX1'
        
    elif indata['c'] == 'tmx3':

        test_type = 'TMX3'
        meter_serno = int(serial)
        
        if indata['x3'] == 'RSM3':
            meter_num_phases = 3
            meter_protocol = 'TMX3'
            
        elif indata['x3'] == 'MC3-24':
            meter_num_phases = 24
            meter_protocol = 'TMX3M'
            
        elif indata['x3'] == 'MC3-48':
            meter_num_phases = 48
            meter_protocol = 'TMX3M'

    elif indata['c'] == 'tmx4':
        test_type = 'TMX4'
        meter_serno = int(serial)
        
    elif indata['c'] == 'tmx5':
        test_type = 'TMX5'
        
    elif indata['c'] == 'tmx5n':
        test_type = 'TMX5n'
        
            

    lw.message('Starting {} test'.format(test_type))

    lw.setaddress(site_ID)

    lw.setserial(serial)

    lw.message('Site ID = {}'.format(site_ID))

    lw.setuser('QLC field technician')
    
    tester = Tester_hardware_obj()

    Radian = Radian_obj("/dev/tty_radian")
    
    if not hasattr(Radian,  'model'):
        lw.message('Unable to contact Radian')
        time.sleep(5)
        lw.message('done')
        return (False)

    lw.message('Radian model {}, Serial Number {},{}, Version {}, Last calibrated {}'.format(Radian.model,
                                                                         Radian.serial,
                                                                         Radian.name,
                                                                         Radian.version,
                                                                         Radian.cal_date))

 #   supply = ACsupply_obj();

    tester.set_volts_on()
    
    lw.volatilemessage('Turning power on ...')

    time.sleep(10)

    Radian.update_readings()

    lw.message('Power on, {:.2f} volts'.format(Radian.volts))

    if Radian.volts < 100.0:

        lw.message('Unable to turn voltage on.  Make sure switch is in test position.')
        
        lw.setfilename('no_file.csv')

        time.sleep(5)
        
        lw.message('done')

        return True
    
    lw.message('Trying to login to meter')
    
    if not os.path.islink('/dev/tty_rs485'):
        port_list = ['/dev/tty_meter']
    elif test_type in ['TMX1','TMX3','TMX4']:
        port_list=['/dev/tty_rs485']
    else:
        port_list=['/dev/tty_meter']
        
    Okay = False
    
    for try_port in port_list:
        
        try:
            DUT = Meter_obj(try_port,
                            test_type,
                            meter_serno,
                            meter_protocol,
                            meter_num_phases)
            Okay = True
            break
        except MeterError as mtr_err:
            hold_message = str(mtr_err)
        except Exception:
            hold_message = 'Script crashed'
            Q_logger.exception('Script failed',exc_info=True)
            break

    if not Okay:
        
        lw.message(hold_message)
        
    else:

        lw.setserial(DUT.meter_serno)

        lw.message('Found meter. Serial number: {}, {} Channels, {}'.format(DUT.meter_serno,
                                                                            DUT.num_phases,
                                                                            DUT.meter_protocol))
        meter_update_delay = 25

##        dump_KWH()

        DUT.channel_mask = DUT.num_phases * [True]

        if DUT.meter_protocol == 'TMX4M':
            lw.message('Found MC-4.  Checking active channels')
            Okay = check_MC_active_phases()

        elif DUT.meter_protocol == 'TMX3M':
            DUT.channel_mask = DUT.num_phases * [False]
            work_message = 'Found MC-3.  Using Channels'
            for work_channel in indata['channels']:
                work_message += ' {}'.format(work_channel)
                DUT.channel_mask[work_channel-1] = True
            lw.message(work_message)
            Okay = DUT.setup_active_channel_scan()

            
                
                
##
##
##              This code will automatically detect active channels on an MC3.
##              Commented out because active channels specified by user.
##
##            lw.message('Found MC-3.  Checking active channels')
##
##            if (DUT.setup_active_channel_scan() and
##                check_MC_active_phases()):
##                Okay = DUT.setup_active_channel_scan()
##            else:
##                Okay = False
                
#            print(DUT.channel_mask)

        if Okay:
            
            tester.preset_amps_low();

            if DUT.use_current_gating:
                tester.suppress_opto_gate()
            else:
                tester.set_amps_on()

            run_datetime = datetime.datetime.now()
            
            year_valid = run_datetime.strftime('%Y')
            
            num_points = 1 if DUT.num_phases <= 3 else DUT.num_phases // 2

            filename = '{}_{}_{}.csv'.format(site_ID,run_datetime.strftime('%Y%m%d_%H%M%S'),
                                          DUT.meter_serno)

            lw.message('Log file name: {}'.format(filename))

            lw.setfilename(filename)
            
            Q_logger.info('Report file: {}/{}'.format(filesloc,filename))

            logfile = open('{}/{}'.format(filesloc,filename),"w")

            logfile.write("Date and time of accuracy tests for,,,,,{},,{},{},test point{}\n".format(site_ID, 
                                                                                                    year_valid, 
                                                                                                    num_points, 
                                                                                                    '' if num_points == 1 else 's'))
            logfile.write(run_datetime.strftime('%m/%d/%Y,,%H:%M:%S'))
            
            logfile.write("\n\nKWH standard used:,,,,,tester #{}\n".format(tester_number))
            logfile.write("Radian Model:,,{}\n".format(Radian.model))
            logfile.write("Serial Number:,,{}\n".format(Radian.serial))
            logfile.write("Version Number:,,{}\n".format(Radian.version))
            logfile.write("Last Calibrated:,,{}\n\n".format(Radian.cal_date))

    ##        test_location = sys.stdin.read()
    ##        logfile.write("type = {}, length = {}\n".format(type(test_location),len(test_location)))
    ##        logfile.write("Test location:,,{}\n".format(test_location))
    ##        lw.message('Test Location: {}'.format(test_location))

    ##        logfile.write("{}".format(dir(sys.stdin)))

            
            logfile.write("Meter under test:\n")
            logfile.write("Serial Number:,,{}\n".format(DUT.meter_serno))
            logfile.write("Software Version:,,{}\n".format(DUT.meter_version))
            logfile.write("# of channels:,,{}\n\n\n\n".format(DUT.num_phases))
            
            tests = Test_obj(tester,
                             Radian,
                             DUT,
                             lw)

            DUT.passed_all_tests = True
            
            DUT.num_active_phases = len([1 for flag in DUT.channel_mask if flag])

            try:

                lw.message('Setting up low-current test.')
                
                tests.do_KWH_test('low')
                
                logfile.write("Accuracy test results:\n")
                
                log_KWH_results('Low Current Test')

                tester.preset_amps_high();
                
                lw.message('Setting up high-current test.')
                
                tests.do_KWH_test('high')

                tester.set_amps_off()

                log_KWH_results('High Current Test')

                work_status = "Pass" if DUT.passed_all_tests else "Fail"

                lw.message("test done.  Meter {}ed.".format(work_status))

                logfile.write("\n\nTest conclusion:,,{}\n\n\n\n".format(work_status))
                
                log_test_details()
                        
                logfile.flush()

            except Exception:

                lw.message('Test not complete.  Problem with meter readings')
                Q_logger.exception('Script failed',exc_info=True)
                logfile.write('Script failed. Problem with meter readings')

                

            logfile.close()

    tester.set_volts_off()

    tester.do_cleanup()

    cmd = "cp -p ./script_ver_work ../script_ver"
    Popen(cmd , shell=True, stdout=PIPE, stderr=PIPE)

    time.sleep(5)

    lw.message('done')

    return True

print(tester_number)
Q_logger = logging.getLogger('QLC_acc_tstr')
Q_logger.setLevel(logging.DEBUG)
lf_hndlr = logging.handlers.RotatingFileHandler('/home/pi/acc_tester/log/acc_test.log',
                                                backupCount=100,
                                                )
lf_hndlr.doRollover()                                               # Force new log file
lf_hndlr.setLevel(logging.DEBUG)
lf_formatter = logging.Formatter('%(asctime)s - %(message)s')
lf_hndlr.setFormatter(lf_formatter)
Q_logger.addHandler(lf_hndlr)                                         # set up logging

for i in range(1):         # test!!!!!!!!!!!!!!!!!!!!!!
    
    Q_logger.info('Starting do_test() on tester# {}'.format(tester_number))

    try:
        s = dotest()
    except Exception:
        Q_logger.exception('Script failed',exc_info=True)
        s = False
    

print("Content-type: text/plain\n")
print("{}".format(s))
sys.stdout.flush()
