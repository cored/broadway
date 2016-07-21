#!/bin/bash

docker-compose up -d
osname=`uname`

# ADD TUNNELING TO K8s
if [[ "$osname" != 'Linux' ]]; then
  dm=$(docker-machine active)
  docker-machine ssh "$dm" -f -N -L "8080:localhost:8080"
  until kubectl -s http://localhost:8080 get pods &> /dev/null; do
    printf "."
  done

  kubectl create -s http://localhost:8080 -f broadway-namespace.yaml
fi

echo "waiting"

