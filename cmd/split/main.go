package main

import (
	"bytes"
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
		msgs = append(msgs, split{raw: raw, Val: hex.EncodeToString(append(splitOn, spl...))})
	}
	log.Printf("sorting by frequency")
	for i := range msgs {
		v := msgs[i]
		msgs[i].Freq = msgset[v.raw]
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
}
