# VERSION = 2.0
#VERSION = 2.01          # Support TMX3
#VERSION = 2.02          # Support incremental tests
#VERSION = 2.03          # Improve logging of test times
VERSION = 2.04          # Log phase diagnostics from Qfam meters

SNAME = "tests.py"

ver_file=open('./script_ver_work','a')
ver_file.write("{} {}\n".format(SNAME,VERSION))
ver_file.close()

from meter import *
from radian import Radian_obj
from ACsupply import ACsupply_obj
from hardware_interface import Tester_hardware_obj
from test_parms import test_parameters, tp_key
import logging
from test_data_log import test_data_log

Q_logger = logging.getLogger('QLC_acc_tstr')


class Test_obj():
    
    def __init__(self,tester,Radian,mtr,logger):
        self.tester = tester
        self.Radian = Radian
        self.mtr = mtr
        self.logger = logger
        self.tester_tol = 0.11
        self.PSC_limit = 2.0
        return

    def _check_pass_fail(self):
        """
        Check test data against pass/fail criteria.  Update log data and files.
        
        Returns:
                    True = results are ambiguous, try longer test time
                    False = results definitely either pass or fail
    
        """

##        work_errs = [(
##            (self.mtr.end_reading[chan_index] -
##             self.mtr.start_reading[chan_index]) /
##            (self.Radian.watt_hours/100) - 1.0) * 100.0 for chan_index in range(
##                len(self.mtr.start_reading)) if self.mtr.channel_mask[chan_index]]

        test_data_log[self.test_point].test_amps = self.Radian.amps * 10
        work_errs =[]
        test_data_log[self.test_point].phase_data = []
        for chan_index in range(len(self.mtr.start_reading)):
            if self.mtr.channel_mask[chan_index]:
                err_temp = (((self.mtr.end_reading[chan_index] - self.mtr.start_reading[chan_index]) /
                            (self.Radian.watt_hours/100) - 1.0) *
                            100.0)
                work_errs += [err_temp]
                test_data_log[self.test_point].phase_data += [(
                                                            self.test_time, 
                                                            self.Radian.watt_hours / 100, 
                                                            self.mtr.start_reading[chan_index], 
                                                            self.mtr.end_reading[chan_index], 
                                                            err_temp)]

        Q_logger.debug('rawerrv: {}\n'.format(' '.join(map('{:.2f}'.format,work_errs))))

#        work_err_total = 0.0
#        
#        if len(work_errs) <= 3:
#            for err in work_errs:
#                work_err_total += err
#            for i in range(len(work_errs)):
#                work_errs[i] = work_err_total / len(work_errs)

        self.mtr.err_percent = [0.0] * self.mtr.num_phases
        self.mtr.raw_err_percent = [0.0] * self.mtr.num_phases
        self.mtr.pass_fail = ['Pass'] * self.mtr.num_phases
        err_index = 0
        found_fail = False
        found_try = False
        log_string = 'ERRv:'

        for chan_index in range(len(self.mtr.start_reading)):
            if self.mtr.channel_mask[chan_index]:
                
                work_percent_err = work_errs[err_index]
                err_index += 1
                
                self.mtr.raw_err_percent[chan_index] = work_percent_err

                if len(work_errs) <= 3:
                   work_percent_err = sum(work_errs) / len(work_errs)
                   
                self.mtr.err_percent[chan_index] = work_percent_err
                log_string += ' {}:{:6.2f}'.format(chan_index+1,work_percent_err)
                
                work_percent_err = abs(work_percent_err)
                
                if (work_percent_err + self.test_uncert + self.tester_tol) < self.PSC_limit:
                    self.mtr.pass_fail[chan_index] = 'Pass'
                elif ((work_percent_err - (self.test_uncert + self.tester_tol) > self.PSC_limit) or
                      self.force_final):
                    self.mtr.pass_fail[chan_index] = 'Fail'
                    self.mtr.passed_all_tests = False
                    found_fail = True
                else:
                    self.mtr.pass_fail[chan_index] = 'Try'
                    found_try = True

        Q_logger.debug(log_string + '\n')
        
        return (not found_fail) and found_try
        

    def _check_pulse_pass_fail(self):
        """
        Check test data against pass/fail criteria.  Update log data and files.
        
        Returns:
                    True = results are ambiguous, try longer test time
                    False = results definitely either pass or fail
    
        """

##        work_errs = [(
##            (self.mtr.end_reading[chan_index] -
##             self.mtr.start_reading[chan_index]) /
##            (self.Radian.watt_hours/100) - 1.0) * 100.0 for chan_index in range(
##                len(self.mtr.start_reading)) if self.mtr.channel_mask[chan_index]]

#        test_data_log[self.test_point].radian_KWH = self.Radian.watt_hours / 100
        test_data_log[self.test_point].test_amps = self.Radian.amps * 10
        work_errs =[]
        test_data_log[self.test_point].phase_data = []
        for chan_index in range(self.mtr.num_phases):
            if self.mtr.channel_mask[chan_index]:
                if self.mtr.pulse_test.current[chan_index].meter == 0:
                    err_temp = -100.0
                else:
                    err_temp = (
                                            (self.mtr.pulse_test.current[chan_index].meter /
                                                ((self.mtr.pulse_test.current[chan_index].radian / 5000) * 
                                                    (2 if self.mtr.meter_protocol == 'QB6' else 3)) - 1.0) *
                                                        100)
                
                work_errs += [err_temp]
                test_data_log[self.test_point].phase_data += [(self.test_time,   # NOTE: try to get actual test time instead of limit
                                                            self.mtr.pulse_test.current[chan_index].radian / 5000, 
                                                            0, 
                                                            self.mtr.pulse_test.current[chan_index].meter, 
                                                            err_temp)]

        Q_logger.debug('rawerrv: {}\n'.format(' '.join(map('{:.2f}'.format,work_errs))))

#        work_err_total = 0.0
#        
#        if len(work_errs) <= 3:
#            for err in work_errs:
#                work_err_total += err
#            for i in range(len(work_errs)):
#                work_errs[i] = work_err_total / len(work_errs)

        self.mtr.err_percent = [0.0] * self.mtr.num_phases
        self.mtr.raw_err_percent = [0.0] * self.mtr.num_phases
        self.mtr.pass_fail = ['Pass'] * self.mtr.num_phases
        err_index = 0
        found_fail = False
        found_try = False
        log_string = 'ERRv:'

        for chan_index in range(self.mtr.num_phases):
            if self.mtr.channel_mask[chan_index]:
                
                work_percent_err = work_errs[err_index]
                err_index += 1
                
                self.mtr.raw_err_percent[chan_index] = work_percent_err

                if len(work_errs) <= 3:
                   work_percent_err = sum(work_errs) / len(work_errs)
                   
                self.mtr.err_percent[chan_index] = work_percent_err
                log_string += ' {}:{:6.2f}'.format(chan_index+1,work_percent_err)
                
                work_percent_err = abs(work_percent_err)
                
                if (work_percent_err + self.test_uncert + self.tester_tol) < self.PSC_limit:
                    self.mtr.pass_fail[chan_index] = 'Pass'
                elif ((work_percent_err - (self.test_uncert + self.tester_tol) > self.PSC_limit) or
                      self.force_final):
                    self.mtr.pass_fail[chan_index] = 'Fail'
                    self.mtr.passed_all_tests = False
                    found_fail = True
                else:
                    self.mtr.pass_fail[chan_index] = 'Try'
                    found_try = True

        Q_logger.debug(log_string + '\n')
        
        return (not found_fail) and found_try
        

    def _setup_test_parms(self):
        
        (self.test_time, self.data_wait, self.test_uncert) = self.test_times[self.time_index]
        test_data_log[self.test_point].test_seconds = self.test_time
        test_data_log[self.test_point].percent_uncertainty = self.test_uncert
        if self.time_index > 0:
            self.test_time -= self.test_times[self.time_index-1][0]
        self.time_index += 1
        if self.time_index == len(self.test_times):
            self.time_index -= 1
            self.force_final = True
        else:
            self.force_final = False
            
        Q_logger.debug('{}: {}, {}: {}, Uncertainty: {}%'.format(
                                            'Timeout Secs' if self.mtr.meter_type == 'QFAM'else 'Accum Secs', 
                                            self.test_time, 
                                            '# Meter Pulses' if self.mtr.meter_type == 'QFAM'else 'Data Wait Secs', 
                                            
                                            self.data_wait, 
                                            self.test_uncert))
        return
    
    def _log_QFAM_diagnostics(self):
        Q_logger.debug('NOTE: For QBrick6, the kW value should be negative')
        Q_logger.debug('meter  CT  Amps   Volts     kW    P.F.')
        for point in self.mtr.pulse_test.meter.point:
            for chnl in point.chan:
                Q_logger.debug('{:^5s} {:^3s} {:^6.3f} {:^7.3f} {:^7.3f} {:^6.2f}'.format(
                                                                                            chnl.point_name,
                                                                                            chnl.chan_name, 
                                                                                            chnl.amps, 
                                                                                            chnl.volts, 
                                                                                            chnl.watts, 
                                                                                            chnl.pf))
        
        pass

    def do_KWH_test(self,test_point):
        """ Passed test_point = ['low'|'high']."""

        self.test_point = test_point
        work_phase_index = 0
        while test_parameters[tp_key[self.mtr.meter_protocol]][work_phase_index][0] < self.mtr.num_active_phases:
            work_phase_index += 1
            
        self.test_times = test_parameters[tp_key[self.mtr.meter_protocol]][work_phase_index][1][self.test_point]
        self.time_index = 0
        self._setup_test_parms()

        if self.mtr.use_current_gating:
            self.time_index = 0                                     # reset index to allow for iteration
            self.tester.set_amps_off()                              # Turn amps off for current gating
            self.mtr.prepare_for_KWH()
            self.Radian.setup_for_gate()
            self.mtr.wait_seconds(self.data_wait,'Meter updating')  # Wait for data to stabilize
            self.tester.do_gate_pulse()                             # Begin Radian accumulation for current gating
            self.mtr.start_reading = self.mtr.get_KWH()             # Get start readings from meter
            not_done = True
            while not_done:
                self._setup_test_parms()
                self.tester.set_amps_on()                                       # Turn on amps to begin KWH accumlation for current gating
                self.mtr.wait_seconds(self.test_time,'Accumulating KWH')
                self.Radian.update_readings()                                   # Get instantaneous values
                self.tester.set_amps_off()                                      # Stop KWH accumulation for current gating - turn amps off
                self.logger.message('Accumulation done, getting data')
                self.mtr.wait_seconds(self.data_wait,'Meter updating')        # Wait for data to stabilize
                self.mtr.send_cmd("\r",cmd_timeout=1,com_retries=5,prompt_re=r"\r\n.*$")      # clear any garbage chars from meter
                self.Radian.read_accumulated()                                  # read metrics
                self.mtr.end_reading = self.mtr.get_KWH()                       # Get end readings from meter
                not_done = self._check_pass_fail()

        else:
            self.tester.set_amps_on()            # Turn amps on for register gating
            if self.mtr.meter_type == 'QFAM':
                self.Radian.start_accumulating()        # Start Radian pulse train
                time.sleep(2)
                self.Radian.update_readings()                                   # Get instantaneous values
                self.mtr.pulse_test.set_meter_pulses()
                self.mtr.pulse_test.start_test(self.data_wait)
                self.mtr.pulse_test.get_phase_diagnostics(request = 'inst')
                self._log_QFAM_diagnostics()                                         # Record phase diagnostics in log                
                self.mtr.wait_pulses(self.test_time, self.data_wait)
                self.mtr.pulse_test.get_test_data()
                Q_logger.info('current values = : {}'.format([(p.meter,p.radian) for p in self.mtr.pulse_test.current]))
                self._check_pulse_pass_fail()
                # start MSP430 pulse test
                # update timer for operator
                # get final results from MSP430
            else:
                self.mtr.freeze_reg()                # Stop accumulation for register gating
                self.mtr.prepare_for_KWH()
                self.Radian.setup_for_gate()
                self.mtr.wait_seconds(self.data_wait,'Meter updating')  # Wait for data to stabilize
                self.mtr.start_reading = self.mtr.get_KWH()         # Get start readings from meter
                self.mtr.arm_reg_gate()                             # Preset meters for register gating
                self.tester.do_gate_pulse()                         # Begin accumulation for register gating
                self.mtr.wait_seconds(self.test_time,'Accumulating KWH')
                self.Radian.update_readings()                       # Get instantaneous values
                self.tester.do_gate_pulse()                         # Stop accumulation for register gating
                self.logger.message('Accumulation done, getting data')
                self.mtr.wait_seconds(self.data_wait,'Meter updating')        # Wait for data to stabilize
                self.mtr.send_cmd("\r",com_retries=5,prompt_re=r"\r\n.*$")      # clear any garbage chars from meter
                self.Radian.read_accumulated()                      # read metrics
                self.mtr.end_reading = self.mtr.get_KWH()           # Get end readings from meter                
                self._check_pass_fail()                             #  Calculate errors and check against limit

        return
            
            

if __name__ == "__main__":
    
    print("start")

    tester = Tester_hardware_obj()

    Radian = Radian_obj("/dev/tty_radian")
    
    print(Radian.model,Radian.serial,Radian.name,Radian.version)

    supply = ACsupply_obj();

    tester.set_volts_on()

    time.sleep(3)
            
    try:
        DUT = Meter_obj("/dev/tty_meter")
        Okay = True
    except MeterError:
        print ("Unable to contact meter")
        Okay = False

    if Okay:

        logfile = open("test_log.txt","w")

#        while True:

        tester.set_amps_low();
        
        print(DUT.meter_type,
              DUT.meter_serno,
              DUT.meter_time,
              DUT.meter_date,
              DUT.meter_version,
              DUT.num_phases,
              DUT.meter_protocol)
                    
        tests = Test_obj()
        hold_readings = tests.do_KWH_test(DUT,30)

        Radian.update_readings()
        print("Test of {} at {:.2f} amps".format(DUT.meter_serno,Radian.amps*10.0))
        logfile.write("{} {:.2f}A\n".format(DUT.meter_serno,Radian.amps*10.0))
    
        tester.set_amps_off()

        for l in hold_readings:
            print("{:6.2f}".format(((l/(Radian.watt_hours/100))-1.0)*100.0))
            logfile.write("{:.2f}\n".format(((l/(Radian.watt_hours/100))-1.0)*100.0))

        tester.set_amps_high();
        
#        print(DUT.meter_type, DUT.meter_serno, DUT.meter_time, DUT.meter_date, DUT.meter_version, DUT.num_phases)
                    
#        tests = Test_obj()
        hold_readings = tests.do_KWH_test(DUT,20)

        Radian.update_readings()
        print("Test of {} at {:.2f} amps".format(DUT.meter_serno,Radian.amps*10.0))
        logfile.write("{} {:.2f}A\n".format(DUT.meter_serno,Radian.amps*10.0))
    
        tester.set_amps_off()

        for l in hold_readings:
            print("{:6.2f}".format(((l/(Radian.watt_hours/100))-1.0)*100.0))
            logfile.write("{:.2f}\n".format(((l/(Radian.watt_hours/100))-1.0)*100.0))

        logfile.flush()

    logfile.close()

    tester.set_volts_off()

    tester.do_cleanup()


