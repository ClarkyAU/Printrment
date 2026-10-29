// Fake printer storage: implicit FTPS on :990 serving ./root, like a Bambu printer.
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const code = "12345678"

var root = "root"
var tcfg *tls.Config

func main() {
	cert, err := tls.LoadX509KeyPair("c.pem", "k.pem")
	if err != nil {
		log.Fatal(err)
	}
	tcfg = &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", "127.0.0.1:990", tcfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mock storage on :990")
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go serve(c)
	}
}

func local(p string) string { return filepath.Join(root, path.Clean("/"+p)) }

func serve(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	say := func(s string) { io.WriteString(c, s+"\r\n") }
	say("220 Bambu FTP ready")
	authed := false
	var pasv net.Listener
	var rest int64
	cwd := "/"
	openData := func() net.Conn {
		if pasv == nil {
			return nil
		}
		d, err := pasv.Accept()
		pasv.Close()
		pasv = nil
		if err != nil {
			return nil
		}
		return d
	}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		cmd, arg, _ := strings.Cut(line, " ")
		cmd = strings.ToUpper(cmd)
		if cmd == "PASS" {
			log.Printf("> PASS ****")
		} else {
			log.Printf("> %s", line)
		}
		if !authed && cmd != "USER" && cmd != "PASS" && cmd != "QUIT" {
			say("530 Please login")
			continue
		}
		p := arg
		if p != "" && !strings.HasPrefix(p, "/") {
			p = path.Join(cwd, p)
		}
		switch cmd {
		case "USER":
			say("331 Password required")
		case "PASS":
			if arg == code {
				authed = true
				say("230 Logged in")
			} else {
				say("530 Login incorrect")
			}
		case "PBSZ", "PROT", "TYPE":
			say("200 OK")
		case "PWD":
			say(`257 "` + cwd + `"`)
		case "CWD":
			cwd = path.Clean("/" + p)
			say("250 OK")
		case "PASV":
			l, err := tls.Listen("tcp", "127.0.0.1:0", tcfg)
			if err != nil {
				say("425 no data")
				continue
			}
			pasv = l
			port := l.Addr().(*net.TCPAddr).Port
			say(fmt.Sprintf("227 Entering Passive Mode (192,168,1,99,%d,%d)", port/256, port%256)) // internal IP on purpose
		case "SIZE":
			st, err := os.Stat(local(p))
			if err != nil || st.IsDir() {
				say("550 No such file")
				continue
			}
			say("213 " + strconv.FormatInt(st.Size(), 10))
		case "REST":
			rest, _ = strconv.ParseInt(arg, 10, 64)
			say("350 Restarting")
		case "LIST":
			dir := cwd
			if p != "" {
				dir = p
			}
			ents, err := os.ReadDir(local(dir))
			if err != nil {
				say("550 No such directory")
				continue
			}
			say("150 Here comes the listing")
			d := openData()
			for _, e := range ents {
				info, _ := e.Info()
				perm := "-rw-r--r--"
				if e.IsDir() {
					perm = "drwxr-xr-x"
				}
				fmt.Fprintf(d, "%s    1 root     root     %8d Oct 28 21:14 %s\r\n", perm, info.Size(), e.Name())
			}
			d.Close()
			say("226 Done")
		case "RETR":
			f, err := os.Open(local(p))
			if err != nil {
				say("550 No such file")
				continue
			}
			say("150 Opening data connection")
			d := openData()
			f.Seek(rest, 0)
			rest = 0
			n, _ := io.Copy(d, f)
			f.Close()
			d.Close()
			log.Printf("  sent %d bytes of %s", n, p)
			say("226 Transfer complete")
		case "ABOR":
			say("226 ABOR ok")
		case "QUIT":
			say("221 Bye")
			return
		default:
			say("502 Not implemented")
		}
	}
}
