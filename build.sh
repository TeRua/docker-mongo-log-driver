#!/bin/bash

PLUGIN_NAME=terua05/mongo-log-driver
PLUGIN_TAG=0.0.2

echo "Cleaning up before start build..."
sudo rm -rf rootfs

echo "Start build ${PLUGIN_NAME}"
docker build -t ${PLUGIN_NAME}:temp .
ID=$(docker create --name plugin-build ${PLUGIN_NAME}:temp true)

echo "Exporting ${PLUGIN_NAME}..."
mkdir rootfs
docker export $ID | tar -x -C rootfs/

echo "Cleaning up build assets..."
docker rm -vf plugin-build

echo "Cleaning up previous plugins..."
docker plugin disable ${PLUGIN_NAME}:${PLUGIN_TAG} || true
docker plugin rm ${PLUGIN_NAME}:${PLUGIN_TAG} || true

echo "Creating docker plugin..."
docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} .
docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

