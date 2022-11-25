class log_data():
    
    def __init__(self):
        
        self.test_amps = 0.0
        self.test_seconds = 0.0
        self.radian_KWH = 0.0
        self.percent_uncertainty = 0.1
        self.phase_data = []
                                # phase_data is a list of tuples (per phase)
                                # (start_reading, end_reading, percent_error)

        return
 
test_data_log = {
                'low' : log_data(), 
                'high' : log_data()
                }
