#!/bin/bash

# These need to be auto-generated in python, like so:
# source /opt/vyatta/etc/functions/script-template
# configure
# output=$(<command> 2>&1) (only with the redirect and capture for rollback detection)
# if [ -n "${output}" ]; then
#    echo "${output}"
#    discard
#    exit 1
# fi
# commit_output=$(commit) # Same note on capturing
# if [ -n "${commit_output}" ]; then
#   echo "${commit_output}"
#   discard
#   exit 1
# fi

# Use like so:
# results=$(ssh <host> ./script.sh)
# exit_code=$?
# if [ $exit_code = 0 ]; then
#   echo "Commands applied successfully
# else
#   echo "${results}"
# fi
