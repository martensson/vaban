package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"regexp"
	"strings"
)

func varnishAuth(server string, secret string, conn net.Conn) error {
	// I want to allocate 512 bytes, enough to read the varnish help output.
	reply := make([]byte, 512)
	conn.Read(reply)
	rp := regexp.MustCompile("[a-z]{32}") //find challenge string
	challenge := rp.FindString(string(reply))
	if challenge != "" {
		// time to authenticate
		hash := sha256.New()
		hash.Write([]byte(challenge + "\n" + secret + "\n" + challenge + "\n"))
		md := hash.Sum(nil)
		mdStr := hex.EncodeToString(md)
		conn.Write([]byte("auth " + mdStr + "\n"))
		auth_reply := make([]byte, 512)
		conn.Read(auth_reply)
		log.Println(server, "auth status", strings.Trim(string(auth_reply)[0:12], " "))
		return nil
	} else {
		return errors.New(server + " no challenge code, secret-file disabled.")
	}
}
