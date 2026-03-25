import logging

from .protocol import parse_agency, parse_batch_count, parse_bet_line
from .utils import winner_documents_by_agency


class ClientHandler:
    def __init__(self, client, submit_bets, on_finished, wait_for_all_finished):
        self._client = client
        self._submit_bets = submit_bets
        self._on_finished = on_finished
        self._wait_for_all_finished = wait_for_all_finished

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
            finished_notified = False
            logging.info(
                f'action: receive_message | result: success | ip: {self._client._ip} | msg: {agency_msg}'
            )

            while True:
                count_msg = self._client.receive_message()
                if count_msg is None:
                    break

                if count_msg.strip() == '':
                    continue

                if count_msg.strip().upper() == 'FIN':
                    logging.info(
                        f'action: receive_message | result: success | ip: {self._client._ip} | agency: {agency} | msg: FIN'
                    )
                    self._on_finished(agency)
                    self._wait_for_all_finished()
                    self._client.send('ok\n')
                    finished_notified = True
                    continue

                if count_msg.strip().upper() == 'WINNERS':
                    if not finished_notified:
                        raise ValueError('winners requested before FIN')

                    winners = winner_documents_by_agency(int(agency))
                    winners_msg = '|'.join(winners) if winners else 'NONE'
                    self._client.send(f'{winners_msg}\n')
                    logging.info(
                        f'action: consulta_ganadores_servidor | result: success | agency: {agency} | cant_ganadores: {len(winners)}'
                    )
                    break

                batch_count = parse_batch_count(count_msg)
                bets = []

                for _ in range(batch_count):
                    raw_bet = self._client.receive_message()
                    if raw_bet is None:
                        raise ValueError('connection closed while receiving batch')

                    bets.append(parse_bet_line(raw_bet, agency))

                self._submit_bets(bets)
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
