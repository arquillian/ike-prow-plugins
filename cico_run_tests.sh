#!/bin/bash

. cico_setup.sh

cico_setup;

trap cleanup_env EXIT;
run_build;
