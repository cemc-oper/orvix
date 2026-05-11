#!/bin/bash
#ORVIX scheduler=slurm
#ORVIX job-name=hello
#ORVIX partition=cpu
#ORVIX nodes=1
#ORVIX ntasks=1
#ORVIX time=00:05:00

echo "Hello from $(hostname) at $(date)"
sleep 2
echo "Done"
