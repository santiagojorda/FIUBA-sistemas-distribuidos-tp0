import socket
import logging
import signal
import threading

from .client import Client
from .client_handler import ClientHandler
from .utils import store_bets, winners_count_by_agency

class Server:
    def __init__(self, port, listen_backlog, amount_clients):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._server_socket.settimeout(1.0)
        self._clients = []
        self._client_threads = []
        self.amount_clients = amount_clients

        self._shutdown_event = threading.Event()
        self._state_lock = threading.Lock()
        self._sorteo_condition = threading.Condition(self._state_lock)
        self._bets_lock = threading.Lock()

        self._finished_agencies = set()
        self._sorteo_done = False
        self._winners_by_agency = {}
        signal.signal(signal.SIGTERM, self.__handle_graceful_shutdown)

    def __handle_graceful_shutdown(self, signum, frame):
        self._shutdown_event.set()
        with self._sorteo_condition:
            self._sorteo_condition.notify_all()
        logging.info('action: graceful_shutdown | result: in_progress')

    def run(self):
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

        for thread in self._client_threads:
            thread.join(timeout=1)

        logging.info('action: graceful_shutdown | result: success')

    def __handle_client_connection(self, client):
        thread = threading.Thread(
            target=self.__run_client_handler,
            args=(client,),
            daemon=True,
        )
        thread.start()
        self._client_threads.append(thread)

    def __run_client_handler(self, client):
        handler = ClientHandler(
            client,
            self.register_finished_agency,
            self.get_winners_count,
            self.store_bets_thread_safe,
        )
        handler.handle()

    def store_bets_thread_safe(self, bets):
        with self._bets_lock:
            store_bets(bets)

    def register_finished_agency(self, agency_id):
        agency = int(agency_id)
        with self._sorteo_condition:
            self._finished_agencies.add(agency)

            if len(self._finished_agencies) >= self.amount_clients and not self._sorteo_done:
                self._winners_by_agency = winners_count_by_agency()
                logging.info('action: sorteo | result: success')
                self._sorteo_done = True
                self._sorteo_condition.notify_all()

    def get_winners_count(self, agency_id):
        agency = int(agency_id)

        with self._sorteo_condition:
            while not self._sorteo_done and not self._shutdown_event.is_set():
                self._sorteo_condition.wait(timeout=1)

            if not self._sorteo_done:
                return None

            return self._winners_by_agency.get(agency, 0)

    def __accept_new_connection(self):
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
            if self._shutdown_event.is_set():
                return None
            logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None