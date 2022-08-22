#!/bin/bash

for i in $(seq 1 $SAMPLES); do
    ./if711-assignment-3-client-udp & 
done