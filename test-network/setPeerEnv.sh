#!/bin/bash

source scriptUtils.sh

# import utils
. scripts/envVar.sh

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
else
  MODE=$1
fi

#set -x
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=$PWD/config/
setGlobals "${MODE}" 

