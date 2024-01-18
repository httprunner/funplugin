import json
import logging
import random
import sys
import time
import socket
from concurrent import futures
from typing import Callable

import grpc

from funppy import debugtalk_pb2, debugtalk_pb2_grpc

__all__ = ["register", "serve"]

functions = {}

def register(func_name: str, func: Callable):
    logging.info(f"register function: {func_name}")
    functions[func_name] = func


class DebugTalkServicer(debugtalk_pb2_grpc.DebugTalkServicer):
    """Implementation of DebugTalk service."""

    def GetNames(self, request: debugtalk_pb2.Empty, context: grpc.ServicerContext):
        names = list(functions.keys())
        response = debugtalk_pb2.GetNamesResponse(names=names)
        return response

    def Call(self, request: debugtalk_pb2.CallRequest, context: grpc.ServicerContext):
        if request.name not in functions:
            raise Exception(f'Function {request.name} not registered!')

        fn = functions[request.name]
        args = json.loads(request.args)
        value = fn(*args)

        if isinstance(value, (int, float)):
            v = str(value).encode('utf-8')
        elif isinstance(value, (str, dict, list)):
            v = json.dumps(value).encode('utf-8')
        else:
            raise Exception(f'Function return type {type(value)} not supported!')

        response = debugtalk_pb2.CallResponse(value=v)
        return response

def check_available_port(port):
    try:
        # Create a socket object and attempt to bind it to the specified port
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(("127.0.0.1", port))
        # The port is available, return False
        return False
    except OSError:
        # The port is already in use, return True
        return True

def serve():
    # Generate a random port
    random_port = random.randrange(20000, 60000)

    # Check if the random port is already in use
    while check_available_port(random_port):
        random_port = random.randrange(20000, 60000)

    # Create the gRPC server and continue with the rest of your code
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    debugtalk_pb2_grpc.add_DebugTalkServicer_to_server(DebugTalkServicer(), server)

    server.add_insecure_port(f'127.0.0.1:{random_port}')
    server.start()

    # Output information
    print(f"1|1|tcp|127.0.0.1:{random_port}|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
