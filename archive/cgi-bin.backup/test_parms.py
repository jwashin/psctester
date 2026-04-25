#VERSION = 2.00      # Initial release
#VERSION = 2.01      # Add support for TMX5D
#VERSION = 2.02      # Add support for TMX4 and other TMX5
VERSION = 2.03      # Correct times for TMX3M

SNAME = 'test_parms.py'

ver_file=open('./script_ver_work','a')
ver_file.write('{} {}\n'.format(SNAME,VERSION))
ver_file.close()

"""
    Description of TP_key:

        This is an alias dictionary for test_parameters.
        The keys are the meter_type
        The values are the key to use in the test_parameters dictionary

"""

    
tp_key = {
          'TMX5n' : 'TMX5n',
          'MC5n'  : 'TMX5n',
          'TMX5B' : 'TMX5',
          'TMX5C' : 'TMX5',
          'TMX5D' : 'TMX5',
          'TMX4'  : 'TMX4',
          'TMX4M' : 'TMX4M',
          'TMX3'  : 'TMX3',
          'TMX3M' : 'TMX3M',
          'TMX1'  : 'TMX1'
          }


"""
Description of the test_parameters object:

    This is a dictionary with the key based on meter_protocol

    The key can have the following values:
        TMX1  - Any family 1 meters - note:  there are no family 1 miniclosets
        TMX3  - Single-point family 3 meters
        TMX3M - Family 3 miniclosets
        TMX4  - Single-point family 4 meters
        TMX4M - Family 4 miniccloset
        TMX5 - All family 5 meters, including family 5 miniclosets
        TMX5n - Family 5n single-point meters
        MC5n - Family 5n miniclosets

    Each entry in the dictionary contains a list with the following elements:
        The number of phases
        A dictionary with the keys 'high' and 'low'
        
            Each entry in these dictionaries contains a list of tuples:
                Each tuple contains the following fields:
                    [0] - test time in seconds
                    [1] - wait time for readings to settle - in seconds
                    [2] - uncertainty of reading when using this test time, in percent
                    
                These tuples are executed in order.  If the meter passes within the implied tolerance,
                the test terminates with 'PASS'.
                If the test fails beyond the uncertainty, the test terminates with 'FAIL'.
                If the test fails within the uncertainty, then the next tuple is executed,
                in order to reduce the uncertainty.
                If the final tuple is executed, the meter either passes or fails depending on that tuple.

"""

test_parameters = {
    'TMX3M' : [
        [1,{'low': [
            (300,10,1.0),           # test seconds, data wait seconds, uncertainty
            (600,10,0.5),
            (1200,10,0.25),
            (3000,10,0.1)
            ],
           'high': [
            (150,10,1.0),           # test seconds, data wait seconds, uncertainty
            (300,10,0.5),
            (600,10,0.25),
            (1500,10,0.1)
            ]
           }],
        [2,{'low': [
            (200,10,1.0),           # test seconds, data wait seconds, uncertainty
            (400,10,0.5),
            (800,10,0.25),
            (1500,10,0.1)
            ],
           'high': [
            (100,10,1.0),           # test seconds, data wait seconds, uncertainty
            (200,10,0.5),
            (400,10,0.25),
            (800,10,0.1)
            ]
           }],
        [3,{'low': [
            (150,10,1.0),           # test seconds, data wait seconds, uncertainty
            (300,10,0.5),
            (600,10,0.25),
            (1000,10,0.1)
            ],
           'high': [
            (90,10,1.0),           # test seconds, data wait seconds, uncertainty
            (180,10,0.5),
            (360,10,0.25),
            (600,10,0.1)
            ]
           }],
        [99,{'low': [
            (300,10,1.0),           # test seconds, data wait seconds, uncertainty
            (600,10,0.5),
            (1200,10,0.25),
            (3000,10,0.1)
            ],
           'high': [
            (150,10,1.0),           # test seconds, data wait seconds, uncertainty
            (300,10,0.5),
            (600,10,0.25),
            (1500,10,0.1)
            ]
           }]
        ],
    'TMX3' : [
        [3,{'low': [
            (150,10,1.0),           # test seconds, data wait seconds, uncertainty
            (300,10,0.5),
            (600,10,0.25),
            (1000,10,0.1)
            ],
           'high': [
            (90,10,1.0),           # test seconds, data wait seconds, uncertainty
            (180,10,0.5),
            (360,10,0.25),
            (600,10,0.1)
            ]
           }]
        ],
    'TMX1' : [
        [3,{'low': [
            (150,10,1.0),           # test seconds, data wait seconds, uncertainty
            (300,10,0.5),
            (600,10,0.25),
            (1000,10,0.1)
            ],
           'high': [
            (90,10,1.0),           # test seconds, data wait seconds, uncertainty
            (180,10,0.5),
            (360,10,0.25),
            (600,10,0.1)
            ]
           }]
        ],
    'TMX4' : [
        [99,{'low': [
            (100,30,0.1)         # Standard value is 100
            ],
           'high': [
            (45,30,0.1)         # standard value is 45
            ]
           }]
        ],
    'TMX4M' : [
        [99,{'low': [
            (200,30,0.1)
            ],
           'high': [
            (150,30,0.1)
            ]
           }]
        ],
    'TMX5' : [
        [1,{'low': [
            (60,5,0.1)
            ],
           'high': [
            (30,5,0.1)
            ]
           }],
        [2,{'low': [
            (60,5,0.1)
            ],
           'high': [
            (30,5,0.1)
            ]
           }],
        [3,{'low': [
            (60,5,0.1)
            ],
           'high': [
            (30,5,0.1)
            ]
           }],
        [99,{'low': [
            (100,15,0.1)
            ],
           'high': [
            (60,15,0.1)
            ]
           }]
        ], 
    'TMX5n': [
        [99,{'low': [
            (240,15,0.25)               # Standard value is 240
            ],
           'high': [
            (240,15,0.25)               # Standard value is 240
            ]
           }]
        ]
    }
        
           
           
        
        
