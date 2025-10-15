def range_lookup(pass_key,  pass_list):
    """
    return value from passed list
    
    passed:
        pass_key = Value to use for lookup
        pass_list = list of lists
            each element is a list with three elements:
                [0] = Lower limit for key value
                [1] = upper limit for key value
                [2] = result to return if key within lower and upper limits
                
    exceptions:
        value_error if key not within any range
        
    """
    for work_list in pass_list:
        if (pass_key >= work_list[0] and pass_key <= work_list[1]):
            return (work_list[2])
    raise ValueError
    
if __name__ == "__main__":
    while True:
        test_val = input('Enter value -->')
        print(range_lookup(test_val, 
                            [
                            ['a', 'b', 'between a and b'], 
                            ['c', 'd', 'between c and d']
                            ]))
