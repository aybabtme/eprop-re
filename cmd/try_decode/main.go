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
	"sort"
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

	var msgs []split
	msgset := make(map[string]int)
	for _, spl := range bytes.Split(bin, splitOn) {
		raw := string(spl)
		if _, ok := msgset[raw]; ok {
			msgset[raw]++
			continue
		}
		msgset[raw] = 1
		msg := append(splitOn, spl...)
		msgs = append(msgs, split{raw: string(msg), Val: hex.EncodeToString(msg)})
	}
	log.Printf("sorting by frequency")
	for i := range msgs {
		v := msgs[i]
		msgs[i].Freq = msgset[v.raw[2:]]
		msgs[i].Parsed = tryParse([]byte(v.raw))

		v = msgs[i]
		log.Printf("freq=%d", v.Freq)
	}
	sort.Slice(msgs, func(i, j int) bool {
		a := msgs[i]
		b := msgs[j]
		if a.Freq == b.Freq {
			if len(a.Val) == len(b.Val) {
				return a.Val < b.Val
			}
			return len(a.Val) < len(b.Val)
		}
		return a.Freq > b.Freq
	})

	log.Printf("encoding results to json")

	if err := json.NewEncoder(os.Stdout).Encode(msgs); err != nil {
		log.Fatal(err)
	}
	log.Printf("all done")
}

type split struct {
	raw  string
	Val  string `json:"val"`
	Freq int    `json:"freq"`

	Parsed *Msg `json:"parse,omitempty"`
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
