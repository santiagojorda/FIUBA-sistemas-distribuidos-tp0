from configparser import ConfigParser
import os


def load_server_config() -> dict:
    """Load server configuration from environment and fallback config.ini."""
    config = ConfigParser(os.environ)
    # If config.ini does not exist, original config object is not modified.
    config.read("config.ini")

    config_params = {}
    try:
        config_params["port"] = int(os.getenv("SERVER_PORT", config["DEFAULT"]["SERVER_PORT"]))
        config_params["listen_backlog"] = int(
            os.getenv("SERVER_LISTEN_BACKLOG", config["DEFAULT"]["SERVER_LISTEN_BACKLOG"])
        )
        config_params["logging_level"] = os.getenv("LOGGING_LEVEL", config["DEFAULT"]["LOGGING_LEVEL"])
        config_params["amount_clients"] = int(os.getenv("AMOUNT_CLIENTS", "1"))
    except KeyError as e:
        raise KeyError(f"Key was not found. Error: {e}. Aborting server")
    except ValueError as e:
        raise ValueError(f"Key could not be parsed. Error: {e}. Aborting server")

    return config_params