import logging

# Every module in this package logs through `logging.getLogger(__name__)`, so all
# records land under this logger. A library must not impose output on the
# application embedding it: the NullHandler keeps Python from falling back to
# "no handlers could be found" when the application configured no logging at all,
# while leaving propagation intact so an application that did configure it still
# receives everything.
logging.getLogger(__name__).addHandler(logging.NullHandler())

__version__ = "1.3.0"

# Version of the GO Feature Flag Provider Specification this provider targets.
# https://gofeatureflag.org/specification/openfeature-provider
#
# The evaluation engine it is pinned to is recorded separately, in
# gofeatureflag_python_provider/wasm/_wasi_version.txt.
__specification_version__ = "1.0"
