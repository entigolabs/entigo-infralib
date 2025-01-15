#!/bin/bash
if [ "$1" == "" ]
then
  echo "Specify the module path as first parameter"
  exit 2
fi
echo "Source code path is $1"
cd $1 || exit 1

MODULETYPE=$(basename $(dirname $1))
MODULENAME=$(basename $1)
echo "Module $MODULENAME type is $MODULETYPE"
SCRIPTPATH=$(dirname "$0")

$SCRIPTPATH/$MODULETYPE/static.sh
if [ $? -ne 0 ]
then 
        echo "$MODULETYPE/$MODULENAME Static tests failed."
        exit 1
fi
echo "$MODULETYPE/$MODULENAME Static tests PASS."

if ls ./test/*_test.go 1>/dev/null 2>&1
then
  $SCRIPTPATH/$MODULETYPE/unit.sh $MODULENAME
  if [ $? -ne 0 ]
  then
          echo "$MODULETYPE/$MODULENAME Unit tests failed."
          exit 2
  fi
  echo "$MODULETYPE/$MODULENAME Unit tests PASS."
else
  echo "$MODULETYPE/$MODULENAME No unit test files found in test folder, skipping unit tests."
fi
