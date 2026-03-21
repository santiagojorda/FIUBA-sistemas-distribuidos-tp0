import socket
import logging
import signal
import threading
import json

from .utils import Bet, store_bets
from .client import Client

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._clients = []
        self._shutdown_event = threading.Event()
        signal.signal(signal.SIGTERM, self.__handle_graceful_shutdown)

    def __handle_graceful_shutdown(self, signum, frame):
        self._shutdown_event.set()
        logging.info('action: graceful_shutdown | result: in_progress')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self._shutdown_event.is_set():
            client_sock = self.__accept_new_connection()
            if client_sock is None:
                continue
            self.__handle_client_connection(client_sock)
        try:
            self._server_socket.close()
        except OSError:
            pass

        for client in list(self._clients):
            try:
                client.close()
            except OSError:
                pass

        logging.info('action: graceful_shutdown | result: success')

    def __handle_client_connection(self, client):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # Primero leer el mensaje de agencia
            agency_msg = client.receive_message()
            if agency_msg is None:
                logging.warning(f'action: receive_message | result: fail | ip: {client._ip} | error: connection closed')
                return
            logging.info(f'action: receive_message | result: success | ip: {client._ip} | msg: {agency_msg}')
            
            # Luego leer los mensajes JSON de los jugadores
            while True:
                player_json = client.receive_json()
                if player_json is None:
                    break

                # Some client env values are being sent with surrounding quotes.
                # Normalize before constructing Bet to avoid parse errors.
                def _clean(value):
                    if isinstance(value, str):
                        return value.strip().strip('"')
                    return value
                
                bet = Bet(
                    agency=_clean(agency_msg),
                    first_name=_clean(player_json['name']),
                    last_name=_clean(player_json['lastname']),
                    document=_clean(player_json['dni']),
                    birthdate=_clean(player_json['birthdate']),
                    number=_clean(player_json['number'])
                )
                
                store_bets([bet])
                logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")

                logging.info(f'action: receive_message | result: success | ip: {client._ip} | bet: {bet}')
                client.send(f"{json.dumps(player_json)}\n")
        except (OSError, json.JSONDecodeError, ValueError, KeyError) as e:
            logging.error(f"action: receive_message | result: fail | ip: {client._ip} | error: {e}")
        finally:
            try:
                client.close()
            except OSError:
                pass

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        try:
            # Connection arrived
            logging.info('action: accept_connections | result: in_progress')
            c, addr = self._server_socket.accept()

            client = Client(addr[0], addr[1], c)
            self._clients.append(client)
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return client
        except socket.timeout:
            return None
        except OSError as e:
            if self._shutdown_event.is_set():
                return None
            logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None
