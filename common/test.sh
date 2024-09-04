#!/bin/bash
if [ "$1" == "" ]
then
  echo "Specify the module path as first parameter"
  exit 2
fi
echo "Source code path is $1"
cd $1 || exit 1

MODULETYPE=$(basename $(dirname $1))
echo "Module type is $MODULETYPE"
SCRIPTPATH=$(dirname "$0")

# $SCRIPTPATH/$MODULETYPE/static.sh
# if [ $? -ne 0 ]
# then 
#         echo "Static tests failed."
#         exit 1
# fi
# echo "Static tests PASS."

if ls ./test/*_test.go 1>/dev/null 2>&1
then
  $SCRIPTPATH/$MODULETYPE/unit.sh
  if [ $? -ne 0 ]
  then
          echo "Unit tests failed."
          exit 2
  fi
  echo "Unit tests PASS."
else
  echo "No unit test files found in test folder, skipping unit tests."
fi
