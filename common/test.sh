#!/bin/bash
MODULETYPE_VERSIONED=$(basename $(dirname $(pwd)))
MODULETYPE=$(echo $(basename $(dirname $(pwd))) | cut -d"-" -f1)
MODULENAME=$(basename $(pwd))
SCRIPTPATH=$(dirname "$0")
$SCRIPTPATH/$MODULETYPE/static.sh "$@"
if [ $? -ne 0 ]
then 
        echo "$MODULETYPE_VERSIONED/$MODULENAME Static tests failed."
        exit 1
fi
echo "$MODULETYPE_VERSIONED/$MODULENAME Static tests PASS."

if ls ./test/*_test.go 1>/dev/null 2>&1
then
  $SCRIPTPATH/$MODULETYPE/unit.sh "$@"
  if [ $? -ne 0 ]
  then
          echo "$MODULETYPE_VERSIONED/$MODULENAME Unit tests failed."
          exit 2
  fi
  echo "$MODULETYPE_VERSIONED/$MODULENAME Unit tests PASS."
else
  echo "$MODULETYPE_VERSIONED/$MODULENAME No unit test files found in test folder, skipping unit tests."
fi
