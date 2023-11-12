from typing import Optional

from pydantic import Field

from gofeatureflag_python_provider.options import BaseModel


class FeatureEvent(BaseModel):
    # Kind for a feature event is feature.
    # A feature event will only be generated if the trackEvents attribute of the flag is set to true.
    kind: str = Field(default="feature", Literal=True)

    # ContextKind is the kind of context which generated an event.
    # This will only be "anonymousUser" for events generated
    # on behalf of an anonymous user or the reserved word "user" for events generated on behalf of a non-anonymous user
    contextKind: str = None

    # UserKey The key of the user object used in a feature flag evaluation.
    # Details for the user object used in a feature
    # flag evaluation as reported by the "feature" event are transmitted periodically with a separate index event.
    userKey: str

    # CreationDate When the feature flag was requested at Unix epoch time in milliseconds.
    creationDate: int

    # Key of the feature flag requested.
    key: str

    # Variation of the flag requested. Flag variation values can be "True", "False", "Default" or "SdkDefault"
    # depending on which value was taken during flag evaluation.
    # "SdkDefault" is used when an error is detected and the
    # default value passed during the call to your variation is used.
    variation: str

    # Value of the feature flag returned by feature flag evaluation.
    value: any

    # Default value is set to true if feature flag evaluation failed, in which case the value returned was the default
    # value passed to variation. If the default field is omitted, it is assumed to be false.
    default: bool

    # Source indicates where the event was generated.
    # This is set to SERVER when the event was evaluated in the relay-proxy
    # and PROVIDER_CACHE when it is evaluated from the cache.
    source: str = Field(default="PROVIDER_CACHE", Literal=True)


class RequestDataCollector(BaseModel):
    # Meta are the extra information added to identify who is calling the endpoint.
    meta: Optional[dict[str, str]] = None

    # Events is the list of the event we send in the payload
    events: list[FeatureEvent] = []
