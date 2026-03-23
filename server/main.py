#!/usr/bin/env python3

import logging
from common.config import load_server_config
from common.logging_config import initialize_log
from common.server import Server


def main():
    config_params = load_server_config()
    logging_level = config_params["logging_level"]
    port = config_params["port"]
    listen_backlog = config_params["listen_backlog"]

    initialize_log(logging_level)

    # Log config parameters at the beginning of the program to verify the configuration
    # of the component
    logging.debug(f"action: config | result: success | port: {port} | "
                  f"listen_backlog: {listen_backlog} | logging_level: {logging_level}")

    # Initialize server and start server loop
    server = Server(port, listen_backlog)
    server.run()


if __name__ == "__main__":
    main()
