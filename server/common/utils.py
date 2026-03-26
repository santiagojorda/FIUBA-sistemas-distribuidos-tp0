import csv
from typing import Iterator
from .bet import Bet

""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])


def winners_count_by_agency() -> dict[int, int]:
    winners: dict[int, int] = {}
    try:
        for bet in load_bets():
            if has_won(bet):
                winners[bet.agency] = winners.get(bet.agency, 0) + 1
    except FileNotFoundError:
        return {}

    return winners

def winner_documents_by_agency(agency: int) -> list[str]:
    winners: list[str] = []
    for bet in load_bets():
        if bet.agency == agency and has_won(bet):
            winners.append(str(bet.document))
    return winners