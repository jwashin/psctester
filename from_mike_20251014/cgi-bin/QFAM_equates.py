## New version for PSC tester

## Bits in status byte

MSG_CHKSUM_ERR	= 0x10
CMD_BUSY 	    = 0x20
MSG_PROC_ERR	= 0x40
BUF_OVRRN_ERR	= 0x80

## Codes in Data ID byte

DID_TEST_STATUS	    = 0x81
DID_TEST_DATA	    = 0x82
DID_LATEST_DATA	    = 0x83
DID_VERSION	        = 0x84

DATA_LENGTH = {                 # Table showing number of bytes in reply data
            0               : 0,
            DID_TEST_STATUS : 34,	# 0x81
            DID_TEST_DATA	: 36,	# 0x82
            DID_LATEST_DATA : 36,	# 0x83
            DID_VERSION	    : 4	    # 0x84
            }


## Command codes in bits 7:4 of command byte

READ_TEST_STATUS    = 0x1
START_TEST	        = 0x2
READ_TEST_DATA      = 0x3
READ_LATEST_DATA    = 0x4
READ_VERSION        = 0x5

## Flag bits in test status bytes

WAITING_FOR_1ST	    = 0x01
TEST_IN_PROGRESS    = 0x02
TEST_FINISHED	    = 0x04
TEST_ERR_RADIAN	    = 0x10

## General limits

I2C_SBUF_LEN    = 7
I2C_RBUF_LEN    = 40

# Defaults

NUM_METERS      = 6                                 # Number of metering points per meter
CHANNELS_PER_METER = 2                      # Number of CT channels per meter
DEFAULT_CENTURY_PREFIX = '20'           # Used to convert 2-digit to 4-digit year

I2C_REFRESH_TIME = 6.0              # Number of seconds before automatic I2C reset

WORK_MODBUS_ADDR = 1            # Modbus address to use for QB comm

## Serial Command parameters

MAX_CMD_LEN     = 40                # Maximum number of characters in command string
COMMAND_TIMEOUT = 6.0               # Number of seconds to wait for char from serial port
TIMEOUT_LIMIT   =   1               # Number of timeouts before resetting port test!!!!!! s/b=100


DEFAULT_CENTURY_PREFIX = '20'       # Used to convert 2-digit to 4-digit year
