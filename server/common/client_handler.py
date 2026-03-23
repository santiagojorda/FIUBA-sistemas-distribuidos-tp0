import logging

from .protocol import parse_agency, parse_batch_count, parse_bet_line
from .utils import store_bets


class ClientHandler:
    def __init__(self, client):
        self._client = client

    def handle(self):
        """
        Handle the full client lifecycle for the bets protocol.
        """
        try:
            agency_msg = self._client.receive_message()
            if agency_msg is None:
                logging.warning(
                    f'action: receive_message | result: fail | ip: {self._client._ip} | error: connection closed'
                )
                return

            agency = parse_agency(agency_msg)
            logging.info(
                f'action: receive_message | result: success | ip: {self._client._ip} | msg: {agency_msg}'
            )

            while True:
                count_msg = self._client.receive_message()
                if count_msg is None:
                    break

                if count_msg.strip() == '':
                    continue

                batch_count = parse_batch_count(count_msg)
                bets = []

                for _ in range(batch_count):
                    raw_bet = self._client.receive_message()
                    if raw_bet is None:
                        raise ValueError('connection closed while receiving batch')

                    bets.append(parse_bet_line(raw_bet, agency))

                store_bets(bets)
                logging.info(
                    f'action: apuesta_recibida | result: success | agency: {agency} | cantidad: {len(bets)}'
                )
                self._client.send('ok\n')
        except (OSError, ValueError, KeyError) as e:
            logging.error(
                f'action: receive_message | result: fail | ip: {self._client._ip} | error: {e}'
            )
        finally:
            try:
                self._client.close()
            except OSError:
                pass
