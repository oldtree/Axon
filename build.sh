#!/bin/bash



function start() {
    docker pull grafana:grafana:latest
    docker pull influxdb:latest
    docker run -d -p 3000:3000  grafana:grafana:latest
    docker run -d -p 8086:8086 influxdb:latest
    nohup ./Axon -c cfg.json &> ./app.log &
}

function build() {
    echo $GOPATH
    echo $GOROOT
    echo $PWD
    go build 
    chmod +X Axon
}


function help() {
    echo "--------------------------help info--------------------------"
    echo "start : start grafana docker and influx docker with localhost ;"
    echo "build : build axonx sever code ;"
}

if [ "$1" == "" ]; then 
    help
elif [ "$1" == "start" ]; then
    start
elif [ "$1" == "build" ]; then
    build
fi
