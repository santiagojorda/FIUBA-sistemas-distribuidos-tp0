import socket
import logging
import signal
import threading
import json

BUFFER = 1024
DECODE = 'utf-8'

class Player:
    def __init__(self, name, lastname, dni, birthdate, number):
        self.name = name
        self.lastname = lastname
        self.dni = dni
        self.birthdate = birthdate
        self.number = number
    
    def __repr__(self):
        return f"Player(name={self.name}, lastname={self.lastname}, dni={self.dni}, birthdate={self.birthdate}, number={self.number})"

class Client:
    def __init__(self, ip, port, sock):
        self._ip = ip
        self._port = port
        self.sock = sock
        self.is_alive = True

    def close(self):
        self.is_alive = False
        if self.sock is not None:
            self.sock.close()
    
    def send(self, msg):
        if self.sock is None:
            return False
        
        data = msg.encode(DECODE) if isinstance(msg, str) else msg
        total_sent = 0
        while total_sent < len(data):
            sent = self.sock.send(data[total_sent:])
            if sent == 0:
                return False  # Conexión cerrada
            total_sent += sent
        return True
        
    
    def receive_message(self):
        """Lee un mensaje de texto (no JSON) hasta encontrar newline"""
        if self.sock is None:
            return None
        
        data = b''
        while True:
            chunk = self.sock.recv(BUFFER)
            if not chunk:
                return None  # Conexión cerrada
            data += chunk
            if b'\n' in data:
                break
        
        return data.rstrip().decode(DECODE)
    
    def receive_json(self):
        """Lee un mensaje JSON hasta encontrar newline"""
        msg = self.receive_message()
        if msg is None:
            return None
        return json.loads(msg)



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
                
                # Crear objeto Player desde el JSON
                player = Player(
                    name=player_json.get('name'),
                    lastname=player_json.get('lastname'),
                    dni=player_json.get('dni'),
                    birthdate=player_json.get('birthdate'),
                    number=player_json.get('number')
                )
                
                logging.info(f'action: receive_message | result: success | ip: {client._ip} | player: {player}')
                client.send(f"{json.dumps(player_json)}\n")
        except (OSError, json.JSONDecodeError) as e:
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
