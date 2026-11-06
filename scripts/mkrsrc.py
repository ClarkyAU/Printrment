#!/usr/bin/env python3
"""Makes cmd/printrment/rsrc_windows_amd64.syso: the Windows resources that make
Printrment.exe look like an ordinary program (icon, version information,
manifest). Go links any .syso it finds next to the code. build.sh runs this.

usage: mkrsrc.py <version, e.g. 0.33> <folder with printrment-<64|128|256>.png> <output .syso>
"""
import struct, sys, os

version, icondir, out = sys.argv[1], sys.argv[2], sys.argv[3]
parts = [int(x) for x in version.split(".")] + [0, 0, 0, 0]
maj, minr, pat, bld = parts[0], parts[1], parts[2], parts[3]
vstr = f"{maj}.{minr}.{pat}.{bld}"

MANIFEST = b"""<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity type="win32" name="Clarky3D.Printrment" version="%s" processorArchitecture="amd64"/>
  <description>Printrment - Bambu Lab printer monitor</description>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
    <security><requestedPrivileges><requestedExecutionLevel level="asInvoker" uiAccess="false"/></requestedPrivileges></security>
  </trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
    <application>
      <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
      <supportedOS Id="{1f676c76-80e1-4239-95bb-83d0f6d0da78}"/>
      <supportedOS Id="{4a2f28e3-53b9-4441-ba9c-d69d4a4a6e38}"/>
    </application>
  </compatibility>
</assembly>
""" % vstr.encode()


def u16(s): return s.encode("utf-16-le")
def pad4(b): return b + b"\0" * (-len(b) % 4)

def vblock(key, val=b"", children=b"", typ=0, vlen=None):
    k = u16(key) + b"\0\0"
    body = pad4(struct.pack("<HHH", 0, 0, 0) + k) + val
    if children:
        body = pad4(body) + children
    vl = len(val) if vlen is None else vlen
    return struct.pack("<HHH", len(body), vl, typ) + body[6:]

def sblock(key, text):
    v = u16(text) + b"\0\0"
    return vblock(key, v, typ=1, vlen=len(text) + 1)

fixed = struct.pack("<13I", 0xFEEF04BD, 0x00010000, (maj << 16) | minr, (pat << 16) | bld,
                    (maj << 16) | minr, (pat << 16) | bld, 0x3F, 0, 0x40004, 1, 0, 0, 0)
strings = [("CompanyName", "Clarky3D"), ("FileDescription", "Printrment - Bambu Lab printer monitor"),
           ("FileVersion", vstr), ("InternalName", "Printrment"),
           ("LegalCopyright", "Copyright (c) Clarky3D"), ("OriginalFilename", "Printrment.exe"),
           ("ProductName", "Printrment"), ("ProductVersion", vstr)]
st = b"".join(pad4(sblock(k, v)) for k, v in strings)
sfi = vblock("StringFileInfo", children=vblock("040904b0", children=st, typ=1), typ=1)
vfi = vblock("VarFileInfo", children=vblock("Translation", struct.pack("<HH", 0x409, 0x4B0)), typ=1)
VERSION = vblock("VS_VERSION_INFO", fixed, pad4(sfi) + pad4(vfi), vlen=len(fixed))

icons = []
for size in (64, 128, 256):
    icons.append((size, open(os.path.join(icondir, f"printrment-{size}.png"), "rb").read()))
grp = struct.pack("<HHH", 0, 1, len(icons))
for i, (size, data) in enumerate(icons, 1):
    grp += struct.pack("<BBBBHHIH", 0 if size >= 256 else size, 0 if size >= 256 else size, 0, 0, 1, 32, len(data), i)

LANG = 0x409
tree = {  # type -> {name -> bytes}
    3: {i: d for i, (_, d) in enumerate(icons, 1)},
    14: {1: grp},
    16: {1: VERSION},
    24: {1: MANIFEST},
}

# ---- lay out the resource directory
def dsize(n): return 16 + 8 * n
offs = {}
pos = dsize(len(tree))
for t in sorted(tree):
    offs[("t", t)] = pos; pos += dsize(len(tree[t]))
for t in sorted(tree):
    for n in sorted(tree[t]):
        offs[("n", t, n)] = pos; pos += dsize(1)
entry_pos = {}
for t in sorted(tree):
    for n in sorted(tree[t]):
        entry_pos[(t, n)] = pos; pos += 16
blob_pos = {}
for t in sorted(tree):
    for n in sorted(tree[t]):
        pos += -pos % 4
        blob_pos[(t, n)] = pos; pos += len(tree[t][n])
total = pos + (-pos % 4)

buf = bytearray(total)
relocs = []
def dirhdr(at, count): struct.pack_into("<IIHHHH", buf, at, 0, 0, 0, 0, 0, count)
dirhdr(0, len(tree))
for i, t in enumerate(sorted(tree)):
    struct.pack_into("<II", buf, 16 + 8 * i, t, 0x80000000 | offs[("t", t)])
    dirhdr(offs[("t", t)], len(tree[t]))
    for j, n in enumerate(sorted(tree[t])):
        struct.pack_into("<II", buf, offs[("t", t)] + 16 + 8 * j, n, 0x80000000 | offs[("n", t, n)])
        dirhdr(offs[("n", t, n)], 1)
        struct.pack_into("<II", buf, offs[("n", t, n)] + 16, LANG, entry_pos[(t, n)])
        data = tree[t][n]
        struct.pack_into("<IIII", buf, entry_pos[(t, n)], blob_pos[(t, n)], len(data), 0, 0)
        relocs.append(entry_pos[(t, n)])
        buf[blob_pos[(t, n)]:blob_pos[(t, n)] + len(data)] = data

# ---- wrap it as a COFF object (what the Go linker reads from a .syso)
raw_ptr = 20 + 40
reloc_ptr = raw_ptr + len(buf)
sym_ptr = reloc_ptr + 10 * len(relocs)
obj = struct.pack("<HHIIIHH", 0x8664, 1, 0, sym_ptr, 1, 0, 0)
obj += struct.pack("<8sIIIIIIHHI", b".rsrc", 0, 0, len(buf), raw_ptr, reloc_ptr, 0, len(relocs), 0, 0x40000040)
obj += bytes(buf)
for r in relocs:
    obj += struct.pack("<IIH", r, 0, 3)  # IMAGE_REL_AMD64_ADDR32NB against the .rsrc symbol
obj += struct.pack("<8sIhHBB", b".rsrc", 0, 1, 0, 3, 0)  # section symbol
obj += struct.pack("<I", 4)  # empty string table
open(out, "wb").write(obj)
print("wrote", out, len(obj), "bytes")
