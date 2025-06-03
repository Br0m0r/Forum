#!/bin/bash

docker build -t forum .

docker run -it --rm -p 8080:8080 forum
