package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	splitOnHex := flag.String("split-on", "2928", "bytes to split on, in hex")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	bin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	splitOn, err := hex.DecodeString(*splitOnHex)
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)

	for _, spl := range bytes.Split(bin, splitOn) {
		spl := append(splitOn, spl...)
		if msg := tryParse(spl); msg != nil {
			enc.Encode(msg)
		}
		log.Printf("%x", spl)
	}
}

type Msg struct {
	Start          uint16
	Unknown1       uint8
	Length         uint8
	Msg            string
	Maybe_checksum uint8
}

func tryParse(raw []byte) *Msg {
	if len(raw) <= 4 {
		log.Printf("too short")
		return nil
	}
	m := &Msg{}
	if raw[0] != 0x29 || raw[1] != 0x28 {
		log.Printf("wrong start bytes %x", raw)
		return nil
	}

	m.Start = binary.LittleEndian.Uint16(raw[0:])
	m.Unknown1 = raw[2]
	m.Length = raw[3]
	if int(m.Length) != len(raw[4:])-1 {
		log.Printf("raw=%x", raw)
		log.Printf("len(raw[3:])-1 == %d, parsed length = %d", len(raw[4:])-1, m.Length)
		return nil
	}
	m.Msg = hex.EncodeToString(raw[4 : 4+m.Length])
	m.Maybe_checksum = raw[4+m.Length]
	return m
}
