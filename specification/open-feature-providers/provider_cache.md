# Provider cache

## Overview
When using an Open Feature provider for GO Feature Flag, the provider will have to call the `relay-proxy` to get the 
value of the evaluation.  
The provider can cache the result of the flag evaluation to avoid having to call the `relay-proxy` on every evaluation.


Cache per reason:


TARGETING_MATCH - cache per user
TARGETING_MATCH_SPLIT - cache per user
SPLIT - cache per user
DISABLED - cache for everyone
DEFAULT - cache per user
STATIC - cache for everyone
UNKNOWN - no cache
ERROR - no cache
OFFLINE - cache for everyone

if has any dynamic flag (experimentation, scheduled, progressive) --> no cache