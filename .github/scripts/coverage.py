#!/usr/bin/env python3
import subprocess
import sys
import shlex


class bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


if len(sys.argv) < 3:
    print("Usage: coverage.py [threshold] [go-coverage-report]")
    sys.exit(1)


threshold = float(sys.argv[1])
report = sys.argv[2]


args = shlex.split(f"go tool cover -func {report}")
p = subprocess.run(args, capture_output=True, text=True)

percent_coverage = float(p.stdout.splitlines()[-1].split()[-1].replace("%", ""))
print(f"{bcolors.BOLD}Coverage: {percent_coverage}%{bcolors.ENDC}")

if percent_coverage < threshold:
    print(f"{bcolors.BOLD}{bcolors.FAIL}Coverage below threshold of {threshold}%{bcolors.ENDC}")
    sys.exit(1)
