// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
'use strict';
var grpc = require('grpc');
var pulumi_codegen_loader_pb = require('../../pulumi/codegen/loader_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

function serialize_pulumirpc_codegen_GetSchemaBytesRequest(arg) {
  if (!(arg instanceof pulumi_codegen_loader_pb.GetSchemaBytesRequest)) {
    throw new Error('Expected argument of type pulumirpc.codegen.GetSchemaBytesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pulumirpc_codegen_GetSchemaBytesRequest(buffer_arg) {
  return pulumi_codegen_loader_pb.GetSchemaBytesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pulumirpc_codegen_GetSchemaBytesResponse(arg) {
  if (!(arg instanceof pulumi_codegen_loader_pb.GetSchemaBytesResponse)) {
    throw new Error('Expected argument of type pulumirpc.codegen.GetSchemaBytesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pulumirpc_codegen_GetSchemaBytesResponse(buffer_arg) {
  return pulumi_codegen_loader_pb.GetSchemaBytesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var LoaderService = exports.LoaderService = {
  getSchemaBytes: {
    path: '/pulumirpc.codegen.Loader/GetSchemaBytes',
    requestStream: false,
    responseStream: false,
    requestType: pulumi_codegen_loader_pb.GetSchemaBytesRequest,
    responseType: pulumi_codegen_loader_pb.GetSchemaBytesResponse,
    requestSerialize: serialize_pulumirpc_codegen_GetSchemaBytesRequest,
    requestDeserialize: deserialize_pulumirpc_codegen_GetSchemaBytesRequest,
    responseSerialize: serialize_pulumirpc_codegen_GetSchemaBytesResponse,
    responseDeserialize: deserialize_pulumirpc_codegen_GetSchemaBytesResponse,
  },
};

exports.LoaderClient = grpc.makeGenericClientConstructor(LoaderService);
