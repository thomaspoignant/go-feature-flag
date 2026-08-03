#!/usr/bin/env python3
"""
Verify that a TinyGo-built wasm binary carries the expected shadow-stack size.

The stack size is set through the custom target JSONs in cmd/wasm/targets/
(`-z stack-size=<bytes>` passed to wasm-ld). That wiring relies on TinyGo's
undocumented target-inheritance merge: if a TinyGo upgrade stops honoring the
flag, the build still succeeds and silently ships a 64KB-stack binary again —
the exact regression behind issue #5651. With `--stack-first` (TinyGo default),
the initial value of the module's `__stack_pointer` global equals the stack
size, which is what this script asserts.

Rather than assuming global 0 is always `__stack_pointer`, the script scans the
global section for every mutable i32 global initialized with `i32.const` and
requires exactly one such candidate (TinyGo's current layout). If TinyGo starts
emitting additional globals, this script fails with an explicit message instead
of checking the wrong one.

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


def skip_init_expr(data: bytes, pos: int) -> int:
    """Skip a WASM constant-expression initializer."""
    while True:
        opcode = data[pos]
        pos += 1
        if opcode == 0x0B:  # end
            return pos
        if opcode == 0x41:  # i32.const
            _, pos = sleb(data, pos)
        elif opcode == 0x42:  # i64.const
            _, pos = sleb(data, pos)
        elif opcode in (0x23, 0x24):  # global.get / global.set
            _, pos = uleb(data, pos)
        else:
            raise SystemExit(
                f"unsupported init-expression opcode 0x{opcode:02x} "
                f"at offset {pos - 1}"
            )


def stack_pointer_candidates(path: str) -> list[tuple[int, int]]:
    """Return (global_index, i32.const init value) for shadow-stack candidates."""
    data = open(path, "rb").read()
    if data[:4] != b"\x00asm":
        raise SystemExit(f"{path}: not a wasm binary")

    candidates: list[tuple[int, int]] = []
    pos = 8
    while pos < len(data):
        section_id = data[pos]
        pos += 1
        size, pos = uleb(data, pos)
        section_start = pos
        if section_id == 6:  # global section
            count, pos = uleb(data, pos)
            for index in range(count):
                value_type = data[pos]
                mutability = data[pos + 1]
                pos += 2
                if value_type != 0x7F or mutability != 0x01:
                    pos = skip_init_expr(data, pos)
                    continue
                if data[pos] != 0x41:  # i32.const
                    pos = skip_init_expr(data, pos)
                    continue
                value, pos = sleb(data, pos + 1)
                if data[pos] != 0x0B:  # end
                    raise SystemExit(
                        f"{path}: global {index} init is not a plain i32.const"
                    )
                pos += 1
                candidates.append((index, value))
            break
        pos = section_start + size
    else:
        raise SystemExit(f"{path}: no global section found")
    return candidates


def stack_pointer_init(path: str) -> int:
    candidates = stack_pointer_candidates(path)
    if len(candidates) == 0:
        raise SystemExit(
            f"{path}: no mutable i32 global with an i32.const initializer found"
        )
    if len(candidates) > 1:
        summary = ", ".join(f"global {idx}={value}" for idx, value in candidates)
        raise SystemExit(
            f"{path}: expected exactly one shadow-stack candidate, found "
            f"{len(candidates)} ({summary}); update verify-wasm-stack.py to "
            "identify __stack_pointer explicitly"
        )
    return candidates[0][1]


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
