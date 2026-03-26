import logging

from .bet_parser import parse_agency, parse_batch_count, parse_bet_line
from .utils import store_bets

MESSAGE_FIN = 'FIN'
MESSAGE_OK = 'ok\n'
MESSAGE_END_BATCH = 'END_BATCH'
MESSAGE_ERROR = 'error\n'

class ClientHandler:
    def __init__(self, client):
        self._client = client

    def get_agency_id(self):
        agency_msg = self._client.receive_message()
        if agency_msg is None:
            raise ValueError('missing agency message')
        
        agency_id = parse_agency(agency_msg)
        logging.info(f'action: receive_message | result: success | ip: {self._client._ip} | agency_id: {agency_id}')
        return agency_id

    def handle(self):
        try:
            agency_id = self.get_agency_id()
            total_bets_received = 0

            while True:
                msg = self._client.receive_message()
                if msg is None:
                    break

                msg = msg.strip()

                if msg == '':
                    continue

                if msg.upper() == MESSAGE_FIN:
                    logging.info(f'action: receive_message | result: success | ip: {self._client._ip} | msg: {MESSAGE_FIN}')
                    logging.info(
                        f'action: apuestas_totales_agencia | result: success | agency_id: {agency_id} | cantidad: {total_bets_received}'
                    )
                    break

                bets = []

                try:
                    bets.append(parse_bet_line(msg, agency_id))

                    while True:
                        raw_bet = self._client.receive_message()
                        if raw_bet is None:
                            raise ValueError('connection closed while receiving batch')
                      
                        raw_bet = raw_bet.strip()
                        if raw_bet == '':
                            continue

                        if raw_bet.upper() == MESSAGE_END_BATCH:
                            break

                        bet = parse_bet_line(raw_bet, agency_id)
                        bets.append(bet)
                    store_bets(bets)
                    total_bets_received += len(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    self._client.send(MESSAGE_OK)
                except (ValueError, KeyError) as batch_error:
                    logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)} | error: {batch_error}')
                    logging.error(f'action: receive_message | result: fail | ip: {self._client._ip} | error: {batch_error}')
                    self._client.send(MESSAGE_ERROR)
        except (OSError, ValueError, KeyError) as e:
            logging.error(f'action: receive_message | result: fail | ip: {self._client._ip} | error: {e}')
        finally:
            try:
                self._client.close()
            except OSError:
                pass