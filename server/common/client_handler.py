import logging

from .bet_parser import parse_agency, parse_batch_count, parse_bet_line
MESSAGE_FIN = 'FIN'
MESSAGE_ASK_WINNERS = 'ASK_WINNERS'
MESSAGE_WAIT = 'wait\n'
MESSAGE_OK = 'ok\n'
MESSAGE_ERROR = 'error\n'

class ClientHandler:
    def __init__(self, client, register_finished_agency, get_winners_count, persist_bets):
        self._client = client
        self._register_finished_agency = register_finished_agency
        self._get_winners_count = get_winners_count
        self._persist_bets = persist_bets

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

                if msg.upper() == MESSAGE_ASK_WINNERS:
                    winners_count = self._get_winners_count(agency_id)
                    if winners_count is None:
                        self._client.send(MESSAGE_WAIT)
                    else:
                        self._client.send(f'{winners_count}\n')
                    break
                
                if msg.upper() == MESSAGE_FIN:
                    logging.info(f'action: receive_message | result: success | ip: {self._client._ip} | msg: {MESSAGE_FIN}')
                    self._register_finished_agency(agency_id)
                    break

                batch_count = parse_batch_count(msg)
                bets = []

                try:
                    for _ in range(batch_count):
                        raw_bet = self._client.receive_message()
                        if raw_bet is None:
                            raise ValueError('connection closed while receiving batch')

                        bet = parse_bet_line(raw_bet, agency_id)
                        bets.append(bet)

                    self._persist_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    self._client.send(MESSAGE_OK)
                except (ValueError, KeyError) as batch_error:
                    logging.info(f'action: apuesta_recibida | result: fail | cantidad: {batch_count}')
                    logging.error(f'action: receive_message | result: fail | ip: {self._client._ip} | error: {batch_error}')
                    self._client.send(MESSAGE_ERROR)
        except (OSError, ValueError, KeyError) as e:
            logging.error(f'action: receive_message | result: fail | ip: {self._client._ip} | error: {e}')
        finally:
            try:
                self._client.close()
            except OSError:
                pass