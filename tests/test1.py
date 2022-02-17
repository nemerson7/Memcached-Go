from pymemcache import Client
import random
import sys

"""
This test case utilizes the pymemcache Python module
Execute this file after the server is running
"""


if __name__ == "__main__":
    print("Running Test 1...")
    if len(sys.argv) < 2:
        print("ERROR: provide ip:port as first arg")
        exit(1)
    addr = sys.argv[1]

    for i in range(100): 
        c = Client(addr)
        teststr = str(random.random())
        c.set("x", teststr, noreply=False)
        result = c.get("x").decode('utf-8')
        assert result == teststr

    print("Tests passed.")
    