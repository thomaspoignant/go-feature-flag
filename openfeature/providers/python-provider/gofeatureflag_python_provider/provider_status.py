from enum import Enum


# ProviderStatus is an enum that represents the status of a provider
class ProviderStatus(Enum):
    NOT_READY = 1
    READY = 2
    STALE = 3
    ERROR = 4
