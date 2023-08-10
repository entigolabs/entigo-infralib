#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
echo "Source code path is $SCRIPTPATH"
cd $SCRIPTPATH || exit 1

if [ -f ./test/k8s_static.sh ]
then
	./test/k8s_static.sh
	if [ $? -ne 0 ]
	then 
		echo "Static tests failed."
		exit 1
	fi
	echo "Static tests PASS."
fi

if [ -f ./test/k8s_unit.sh ]
then
        ./test/k8s_unit.sh
        if [ $? -ne 0 ]
        then
                echo "Unit tests failed."
                exit 2
        fi
        echo "Unit tests PASS."
fi
