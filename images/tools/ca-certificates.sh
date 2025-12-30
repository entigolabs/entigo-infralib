#!/bin/bash
if [ -d "ca-certificates" ]
then
  echo "Additional CA certificates found, running update-ca-certificates"
  cp ca-certificates/*.crt /usr/local/share/ca-certificates/
  update-ca-certificates
fi
