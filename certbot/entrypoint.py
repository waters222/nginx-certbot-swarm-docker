import os
import re
import time
import signal
import sys


def signal_handler(signal, frame):
    print "INFO: Exit Certbot Python Script"
    sys.exit(0)


print "Certbot Python Script v0.1"

if not os.environ.has_key("DOMAINS"):
        print "ERROR: no domains defined"
        exit(1)

temp_domain_list = os.environ['DOMAINS'].split(',')

prog = re.compile("^[a-z0-9]([a-z0-9-]+\.){1,}[a-z0-9]+\Z$")

domains = []
for d in temp_domain_list:
    d = d.strip()
    if prog.match(d):
        domains.append(d)
    else:
        print "WARN: "+d+" is not valid domain name"
if len(domains) == 0:
    print "ERROR: we have no domains to process"
    exit(1)
print "INFO: start processing domains: "+', '.join(domains)

signal.signal(signal.SIGINT, signal_handler)

while True:
    time.sleep(60)


