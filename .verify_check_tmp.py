import importlib.util
spec = importlib.util.spec_from_file_location('verify_wasm_stack', '.github/ci-scripts/verify-wasm-stack.py')
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)


def uleb(n):
    out = bytearray()
    while True:
        b = n & 0x7F
        n >>= 7
        if n:
            out.append(b | 0x80)
        else:
            out.append(b)
            return bytes(out)


def sleb(n):
    out = bytearray()
    more = True
    while more:
        b = n & 0x7F
        n >>= 7
        if (n == 0 and not (b & 0x40)) or (n == -1 and (b & 0x40)):
            more = False
        else:
            b |= 0x80
        out.append(b)
    return bytes(out)


def build(globals_spec, extra_sections=b""):
    body = uleb(len(globals_spec))
    for value_type, mutability, opcode, value in globals_spec:
        body += bytes([value_type, mutability, opcode]) + sleb(value) + bytes([0x0B])
    section = bytes([6]) + uleb(len(body)) + body
    return b"\x00asm" + b"\x01\x00\x00\x00" + extra_sections + section


tmp = "/home/runner/work/go-feature-flag/go-feature-flag/.tmp_wasm_test_"

data = build([(0x7F, 0x01, 0x41, 65536)])
open(tmp + "1.wasm", "wb").write(data)
print("single:", m.stack_pointer_init(tmp + "1.wasm"))

data = build([(0x7F, 0x00, 0x41, 65536)])
open(tmp + "2.wasm", "wb").write(data)
try:
    m.stack_pointer_init(tmp + "2.wasm")
    print("no exception! BUG")
except SystemExit as e:
    print("zero candidates:", e)

data = build([(0x7F, 0x01, 0x41, 65536), (0x7F, 0x01, 0x41, 32768)])
open(tmp + "3.wasm", "wb").write(data)
try:
    m.stack_pointer_init(tmp + "3.wasm")
    print("no exception! BUG")
except SystemExit as e:
    print("two candidates:", e)

other_section = bytes([1]) + uleb(2) + b"\x00\x00"
data = build([(0x7F, 0x01, 0x41, 65536)], extra_sections=other_section)
open(tmp + "4.wasm", "wb").write(data)
print("with preceding section:", m.stack_pointer_init(tmp + "4.wasm"))

data = b"\x00asm" + b"\x01\x00\x00\x00" + other_section
open(tmp + "5.wasm", "wb").write(data)
try:
    m.stack_pointer_init(tmp + "5.wasm")
    print("no exception! BUG")
except SystemExit as e:
    print("no global section:", e)

body = uleb(1) + bytes([0x7F, 0x01, 0x41]) + sleb(65536) + bytes([0x01])
section = bytes([6]) + uleb(len(body)) + body
data = b"\x00asm" + b"\x01\x00\x00\x00" + section
open(tmp + "6.wasm", "wb").write(data)
try:
    m.stack_pointer_init(tmp + "6.wasm")
    print("no exception! BUG")
except SystemExit as e:
    print("malformed init:", e)
