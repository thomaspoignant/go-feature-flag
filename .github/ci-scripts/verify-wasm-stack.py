#!/usr/bin/env python3
"""
Verify that a TinyGo-built wasm binary carries the expected shadow-stack size.

The stack size is set through the custom target JSONs in cmd/wasm/targets/
(`-z stack-size=<bytes>` passed to wasm-ld). That wiring relies on TinyGo's
undocumented target-inheritance merge: if a TinyGo upgrade stops honoring the
flag, the build still succeeds and silently ships a 64KB-stack binary again —
the exact regression behind issue #5651. With `--stack-first` (TinyGo default),
the initial value of the module's global 0 (`__stack_pointer`) equals the
stack size, which is what this script asserts.

Usage: verify-wasm-stack.py <binary.wasm> <expected-stack-bytes>
"""

import sys


def uleb(data: bytes, pos: int) -> tuple[int, int]:
    result = shift = 0
    while True:
        byte = data[pos]
        pos += 1
        result |= (byte & 0x7F) << shift
        if not byte & 0x80:
            return result, pos
        shift += 7


def sleb(data: bytes, pos: int) -> tuple[int, int]:
    result = shift = 0
    while True:
        byte = data[pos]
        pos += 1
        result |= (byte & 0x7F) << shift
        shift += 7
        if not byte & 0x80:
            if byte & 0x40:
                result -= 1 << shift
            return result, pos


def stack_pointer_init(path: str) -> int:
    data = open(path, "rb").read()
    if data[:4] != b"\x00asm":
        raise SystemExit(f"{path}: not a wasm binary")
    pos = 8
    while pos < len(data):
        section_id = data[pos]
        pos += 1
        size, pos = uleb(data, pos)
        if section_id == 6:  # global section
            body_pos = pos
            _, body_pos = uleb(data, body_pos)  # global count
            body_pos += 2  # value type + mutability of global 0
            if data[body_pos] != 0x41:  # i32.const
                raise SystemExit(f"{path}: global 0 is not an i32.const")
            value, _ = sleb(data, body_pos + 1)
            return value
        pos += size
    raise SystemExit(f"{path}: no global section found")


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    path, expected = sys.argv[1], int(sys.argv[2])
    actual = stack_pointer_init(path)
    if actual != expected:
        raise SystemExit(
            f"{path}: shadow stack is {actual} bytes (0x{actual:x}), "
            f"expected {expected} (0x{expected:x}) — the `-z stack-size` "
            "linker flag from cmd/wasm/targets/*.json was not applied"
        )
    print(f"{path}: shadow stack OK ({actual} bytes)")


if __name__ == "__main__":
    main()
