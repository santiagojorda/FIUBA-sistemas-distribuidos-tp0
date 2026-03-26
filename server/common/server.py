import socket
import logging
import signal

from .client import Client
from .client_handler import ClientHandler
from .utils import winners_count_by_agency

class Server:
    def __init__(self, port, listen_backlog, amount_clients):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._clients = []
        self.amount_clients = amount_clients
        self._shutdown = False
        self._finished_agencies = set()
        self._sorteo_done = False
        self._winners_by_agency = {}
        signal.signal(signal.SIGTERM, self.__handle_graceful_shutdown)

    def __handle_graceful_shutdown(self, signum, frame):
        self._shutdown = True
        logging.info('action: graceful_shutdown | result: in_progress')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self._shutdown:
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
        handler = ClientHandler(client, self.register_finished_agency, self.get_winners_count)
        handler.handle()

    def register_finished_agency(self, agency_id):
        agency = int(agency_id)
        self._finished_agencies.add(agency)

        if len(self._finished_agencies) >= self.amount_clients and not self._sorteo_done:
            self._winners_by_agency = winners_count_by_agency()
            logging.info('action: sorteo | result: success')
            self._sorteo_done = True

    def get_winners_count(self, agency_id):
        if not self._sorteo_done:
            return None

        agency = int(agency_id)
        return self._winners_by_agency.get(agency, 0)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        try:
            logging.info('action: accept_connections | result: in_progress')
            client_socket, addr = self._server_socket.accept()

            client = Client(addr[0], addr[1], client_socket)
            self._clients.append(client)
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return client
        except socket.timeout:
            return None
        except OSError as e:
            if self._shutdown:
                return None
            logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None