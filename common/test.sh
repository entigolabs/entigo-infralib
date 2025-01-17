#!/bin/bash
MODULETYPE=$(basename $(dirname $(pwd)))
MODULENAME=$(basename $(pwd))
SCRIPTPATH=$(dirname "$0")
$SCRIPTPATH/$MODULETYPE/static.sh "$@"
if [ $? -ne 0 ]
then 
        echo "$MODULETYPE/$MODULENAME Static tests failed."
        exit 1
fi
echo "$MODULETYPE/$MODULENAME Static tests PASS."

if ls ./test/*_test.go 1>/dev/null 2>&1
then
  $SCRIPTPATH/$MODULETYPE/unit.sh "$@"
  if [ $? -ne 0 ]
  then
          echo "$MODULETYPE/$MODULENAME Unit tests failed."
          exit 2
  fi
  echo "$MODULETYPE/$MODULENAME Unit tests PASS."
else
  echo "$MODULETYPE/$MODULENAME No unit test files found in test folder, skipping unit tests."
fi
