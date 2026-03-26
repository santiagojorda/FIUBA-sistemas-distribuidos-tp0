import logging

from .parser import parse_agency, parse_bet_line
from .utils import store_bets

MESSAGE_FIN = 'FIN'

class ClientHandler:
    def __init__(self, client):
        self._client = client

    def get_agency_id(self):
        """
        Get the agency id from the client.

        The client must send a message with the agency id. If the message is empty or None, an error is raised.
        """
        agency_msg = self._client.receive_message()
        if agency_msg is None:
            raise ValueError('missing agency message')
        
        agency_id = parse_agency(agency_msg)
        logging.info(
            f'action: receive_message | result: success | ip: {self._client._ip} | agency_id: {agency_id}'
        )
        return agency_id

    def handle(self):
        """
        Handle the full client lifecycle for the bets protocol.
        """
        try:
            agency_id = self.get_agency_id()

            while True:
                msg = self._client.receive_message()
                if msg is None:
                    break

                msg = msg.strip()

                if msg == '':
                    continue

                if msg.upper() == MESSAGE_FIN:
                    logging.info(f'action: receive_message | result: success | ip: {self._client._ip} | msg: {MESSAGE_FIN}')
                    break

                bet = parse_bet_line(msg, agency_id)
                store_bets([bet])
                logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
                self._client.send('ok\n')
        except (OSError, ValueError, KeyError) as e:
            logging.error(f'action: receive_message | result: fail | ip: {self._client._ip} | error: {e}')
        finally:
            try:
                self._client.close()
            except OSError:
                pass