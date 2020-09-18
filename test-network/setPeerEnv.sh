#!/bin/bash

source scriptUtils.sh

# import utils
. scripts/envVar.sh

## Set Variables for Org 
infoln "Setting Environment Varibales for Org ${MODE}"

# Print the usage message
function printHelp() {
  println "Usage: "
  println "  setPeerEnv.sh <Org#> "
  println " Examples:"
  println "  setPeerEnv.sh 1 "
}

## Parse mode
if [[ $# -lt 1 ]] ; then
  printHelp
  exit 0
else
  MODE=$1
fi

set -x
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=$PWD/config/
setGlobals "${MODE}" 

exit 0
