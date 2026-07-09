package main

import (
	"bytes"
	"log"
	"net"
	"os"

	"github.com/mdlayher/wol"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	matchMAC := mustEnv("WOLPROXY_MATCH_MAC")
	targetMAC := mustEnv("WOLPROXY_TARGET_MAC")
	iface := mustEnv("WOLPROXY_IFACE")
	listenAddr := envOr("WOLPROXY_LISTEN", ":9")

	wifiTarget, err := net.ParseMAC(matchMAC)
	if err != nil {
		log.Fatalf("parsing WOLPROXY_MATCH_MAC %q: %v", matchMAC, err)
	}

	ethTarget, err := net.ParseMAC(targetMAC)
	if err != nil {
		log.Fatalf("parsing WOLPROXY_TARGET_MAC %q: %v", targetMAC, err)
	}

	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		log.Fatalf("looking up interface %q: %v", iface, err)
	}

	raw, err := wol.NewRawClient(ifi)
	if err != nil {
		log.Fatalf("opening raw socket on %q: %v", iface, err)
	}
	defer raw.Close()

	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatalf("listening on UDP %q: %v", listenAddr, err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}
		log.Printf("got UDP broadcast from %s, len=%d", addr, n)

		var mp wol.MagicPacket
		if err := mp.UnmarshalBinary(buf[:n]); err != nil {
			log.Printf("not a magic packet, ignoring: %v", err)
			continue
		}
		log.Printf("parsed target MAC: %s", mp.Target)

		if !bytes.Equal(mp.Target, wifiTarget) {
			log.Print("no match, ignoring.")
			continue
		}

		log.Print("match! Sending raw frame...")
		if err := raw.Wake(ethTarget); err != nil {
			log.Printf("sending magic packet failed: %v", err)
			continue
		}
		log.Print("sent.")
	}
}
