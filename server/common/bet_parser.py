from .bet import Bet

def parse_agency(raw_agency: str) -> str:
    if raw_agency is None:
        raise ValueError('missing agency')

    agency = raw_agency.strip().strip('"')
    if agency == '':
        raise ValueError('empty agency')

    return agency


def parse_batch_count(raw_count: str) -> int:
    if raw_count is None:
        raise ValueError('missing batch count')

    count_msg = raw_count.strip()
    if count_msg == '':
        raise ValueError('empty batch count')

    batch_count = int(count_msg)
    if batch_count < 0:
        raise ValueError('batch count must be positive')

    return batch_count


def parse_bet_line(raw_bet: str, agency: str) -> Bet:
    if raw_bet is None:
        raise ValueError('missing bet data')

    fields = [field.strip().strip('"') for field in raw_bet.split('|')]
    if len(fields) != 5:
        raise ValueError(f'invalid bet format: {raw_bet}')

    return Bet(
        agency=agency,
        first_name=fields[0],
        last_name=fields[1],
        document=fields[2],
        birthdate=fields[3],
        number=fields[4],
    )