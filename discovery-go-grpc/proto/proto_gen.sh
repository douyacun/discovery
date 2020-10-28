#!/usr/bin/env bash

CurDir=`pwd`
ProtoPath=$(cd $(dirname $0); pwd)
ProjectPath=$(cd "../../$(dirname $0)"; pwd)
echo "==============help message=============="
echo ">> current path: ${CurDir}"
echo ">> proto path: ${ProtoPath}"
echo ">> project path: ${ProjectPath}"
echo "==============start build==============="

echo ">> protoc ${ProtoPath}/*.proto --proto_path=${ProtoPath} --proto_path=${ProjectPath} --go_out=plugins=grpc:${ProtoPath}"

protoc ${ProtoPath}/*.proto --proto_path=${ProtoPath} --proto_path=${ProjectPath} --go_out=plugins=grpc:${ProtoPath}
