import socket
import logging
import signal
import threading

from .client import Client
from .client_handler import ClientHandler

class Server:
    def __init__(self, port, listen_backlog, amount_clients):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._clients = []
        self._client_threads = []
        self._shutdown_event = threading.Event()
        self._amount_clients = amount_clients
        self._finished_agencies = set()
        self._finish_condition = threading.Condition()
        self._sorteo_logged = False
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

        for thread in list(self._client_threads):
            thread.join()

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
        handler = ClientHandler(
            client,
            self._register_finished_agency,
            self._wait_for_all_finished,
        )
        thread = threading.Thread(target=handler.handle)
        thread.start()
        self._client_threads.append(thread)

    def _register_finished_agency(self, agency):
        with self._finish_condition:
            self._finished_agencies.add(agency)
            if len(self._finished_agencies) >= self._amount_clients:
                if not self._sorteo_logged:
                    logging.info('action: sorteo | result: success')
                    self._sorteo_logged = True
                self._finish_condition.notify_all()

    def _wait_for_all_finished(self):
        with self._finish_condition:
            while len(self._finished_agencies) < self._amount_clients and not self._shutdown_event.is_set():
                self._finish_condition.wait(timeout=1.0)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """


        # una vez que el servidor no le llegan mas peticiones,
        # se queda a la escucha de que le lleguen todos las apuestas
        # cada batch de apuesta tiene un delimitador que le indica al servidor
        # cuantas apuestas estan siendo enviadas
        #  
        # hace el sorteo

        try:
            # Connection arrived
            logging.info('action: accept_connections | result: in_progress')
            client_socket, addr = self._server_socket.accept()

            client = Client(addr[0], addr[1], client_socket)
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
