import struct
from pymodbus.client import ModbusSerialClient as ModbusClient
from pymodbus.pdu import ModbusRequest
import time


class SetModbus(ModbusRequest):
    function_code = 0x42
    _rtu_frame_size = 13

    def __init__(self, action, timeout, true_addr, not_addr, match_addr, and_mask, or_mask, **kwargs):
        kwargs['unit'] = 0
        ModbusRequest.__init__(self, **kwargs)
        self.action = action
        self.timeout = timeout
        self.true_addr = true_addr
        self.not_addr = not_addr
        self.match_addr = match_addr
        self.and_mask = and_mask
        self.or_mask = or_mask

    def encode(self):
        return struct.pack('>BBBBBII', self.action, self.timeout, self.true_addr, self.not_addr, self.match_addr, self.and_mask, self.or_mask)

    def decode(self, data):
        self.action, self.timeout, self.true_addr, self.not_addr, self.match_addr, self.and_mask, self.or_mask = struct.unpack('>BBBBBII', data)

    def execute(self, context):
        values = context.getValues(self.function_code, self.action, self.timeout, self.true_addr, self.not_addr, self.match_addr, self.and_mask, self.or_mask)
        return values
        
class set_modbus_addr_obj():

    def __init__(self, modbus_client):
        
        self.modbus_client = modbus_client  # Modbus client to use for communication
        self.action = 1                     # Set all command parameters to default
        self.timeout = 1
        self.true_addr = 0
        self.not_addr = 0
        self.match_addr = 0
        self.and_mask = 0x00000000
        self.or_mask = 0xffffffff
        
    def send(self, action = None,   # Allow override on any parameter, value unchanged if parameter not specified
                    timeout = None, 
                    true_addr = None, 
                    not_addr = None, 
                    match_addr = None, 
                    and_mask = None, 
                    or_mask = None
                    ):
                        
        if action != None:
            self.action = action
        if timeout != None:
            self.timeout = timeout
        if true_addr != None:
            self.true_addr = true_addr
        if not_addr != None:
            self.not_addr = not_addr
        if match_addr != None:
            self.match_addr = match_addr
        if and_mask != None:
            self.and_mask = and_mask
        if or_mask != None:
            self.or_mask = or_mask
            
        # Broadcast function to set modbus address on QBrick meters
        result = self.modbus_client.execute(SetModbus(self.action, 
                                                        self.timeout, 
                                                        self.true_addr, 
                                                        self.not_addr, 
                                                        self.match_addr, 
                                                        self.and_mask, 
                                                        self.or_mask))

        time.sleep(0.03)
        return (result)


if (__name__ == '__main__'):
    
    port_addr = input('Enter name of serial device to use for modbus --->')
    
    use_addr = 1
    start_reg = 1328
    num_regs = 3

    modbus_client = ModbusClient(method='rtu',
                                 port='/dev/ttySIP1',
                                 timeout=1,
                                 broadcast_enable = True, 
                                 baudrate=9600)
                                 
    modbus_masks = set_modbus_addr_obj(modbus_client)
    
    # Broadcast function to set modbus address on QBrick meters 
    #   - set all meters to address use_addr
    
    result = modbus_masks.send(action = 1, 
                        timeout = 1, 
                        true_addr = use_addr, 
                        not_addr = 0, 
                        match_addr = 0, 
                        and_mask = 0x00000000, 
                        or_mask = 0xFFFFFFFF)
                        
    print('66 command result = {}'.format(result))
                        
    try:
        result = modbus_client.read_holding_registers(start_reg, unit=use_addr, count=num_regs)
    except Exception as e:
        print('Exception while reading registers: port = {}, {}'.format(modbus_client.port, e))
        result = None

    if result:
        if result.isError():
            print('{}'.format(result))
            result = None
        else:
            RegNo = start_reg
            for reg_val in result.registers:
                print('Regno = {}, RegVal = {:04x}'.format(RegNo, reg_val))
                RegNo += 1
