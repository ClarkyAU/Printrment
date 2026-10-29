import socket, ssl, struct, time, base64, hashlib, re, threading, os, select

def log(*a): print(time.strftime("%H:%M:%S"), *a, flush=True)

data = open("t.h264", "rb").read()
nals = [n for n in re.split(b"\x00\x00\x00\x01|\x00\x00\x01", data) if n]
sps = [n for n in nals if n[0] & 31 == 7][0]
pps = [n for n in nals if n[0] & 31 == 8][0]
sdp = ("v=0\r\nm=video 0 RTP/AVP 96\r\na=control:streaming/live/1\r\na=rtpmap:96 H264/90000\r\n"
       "a=fmtp:96 packetization-mode=1;sprop-parameter-sets=%s,%s\r\n" % (
           base64.b64encode(sps).decode(), base64.b64encode(pps).decode())).encode()
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain("c.pem", "k.pem")
NONCE, CODE = "abc123", "12345678"
DROP = float(os.environ.get("DROP_AFTER", "0"))


def read_req(f):
    line = f.readline()
    if not line:
        return None
    m, uri, _ = line.decode().split(" ", 2)
    h = {}
    while True:
        l = f.readline().decode().strip()
        if not l:
            break
        k, v = l.split(":", 1)
        h[k.lower()] = v.strip()
    return m, uri, h


def serve(c):
    try:
        t = ctx.wrap_socket(c, server_side=True)
        f = t.makefile("rb")
        while True:
            req = read_req(f)
            if req is None:
                log("client closed (no TEARDOWN)")
                return
            m, uri, h = req
            log("REQ", m)
            cs = h["cseq"]

            def reply(code, txt, extra="", body=b""):
                r = "RTSP/1.0 %d %s\r\nCSeq: %s\r\n%s" % (code, txt, cs, extra)
                if body:
                    r += "Content-Length: %d\r\n" % len(body)
                t.sendall(r.encode() + b"\r\n" + body)

            a = h.get("authorization", "")
            ok = False
            if a.startswith("Digest"):
                p = dict(re.findall(r'(\w+)="([^"]*)"', a))
                ha1 = hashlib.md5(("bblp:IPCAM:%s" % CODE).encode()).hexdigest()
                ha2 = hashlib.md5(("%s:%s" % (m, p["uri"])).encode()).hexdigest()
                ok = p.get("response") == hashlib.md5(("%s:%s:%s" % (ha1, NONCE, ha2)).encode()).hexdigest()
            if not ok:
                reply(401, "Unauthorized", 'WWW-Authenticate: Digest realm="IPCAM", nonce="%s"\r\n' % NONCE)
                continue
            if m == "OPTIONS":
                reply(200, "OK", "Public: OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN\r\n")
            elif m == "DESCRIBE":
                reply(200, "OK", "Content-Type: application/sdp\r\n", sdp)
            elif m == "SETUP":
                reply(200, "OK", "Session: 77;timeout=60\r\nTransport: RTP/AVP/TCP;unicast;interleaved=0-1\r\n")
            elif m == "TEARDOWN":
                reply(200, "OK")
                log("got TEARDOWN - session freed")
                return
            elif m == "PLAY":
                reply(200, "OK", "Session: 77\r\n")
                seq = ts = 0
                st = time.time()
                while True:
                    for n in nals[:60]:
                        if DROP and time.time() - st > DROP:
                            log("MOCK drops connection")
                            t.close()
                            return
                        pk = [n] if len(n) <= 1400 else []
                        if not pk:
                            body, hd = n[1:], n[0] & 0xe0
                            for i in range(0, len(body), 1400):
                                fh = (n[0] & 31) | (0x80 if i == 0 else 0) | (0x40 if i + 1400 >= len(body) else 0)
                                pk.append(bytes([hd | 28, fh]) + body[i:i + 1400])
                        for p in pk:
                            rtp = struct.pack(">BBHII", 0x80, 96, seq & 0xffff, ts & 0xffffffff, 1) + p
                            seq += 1
                            t.sendall(b"$" + struct.pack(">BH", 0, len(rtp)) + rtp)
                        if n[0] & 31 in (1, 5):
                            ts += 6000
                            time.sleep(1 / 15)
                        r, _, _ = select.select([t], [], [], 0)
                        if r or t.pending():
                            req = read_req(f)
                            if req is None:
                                log("client closed (no TEARDOWN)")
                                return
                            log("REQ-during-play", req[0])
                            if req[0] == "TEARDOWN":
                                log("got TEARDOWN - session freed")
                                return
    except Exception as e:
        log("mock:", repr(e))


s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("127.0.0.1", 322))
s.listen(5)
while True:
    c, _ = s.accept()
    log("CONNECT")
    threading.Thread(target=serve, args=(c,), daemon=True).start()
