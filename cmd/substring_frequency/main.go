package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	bin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	minLen := flag.Int("min-len", 2, "minimum length of bytes to check for")
	maxLen := flag.Int("max-len", 1024, "maximum length of bytes to check for")
	minFreq := flag.Int("min-freq", 3, "minimum number of frequency in the document to be counted")
	flag.Parse()

	log.Printf("preparing suffixarray")
	sa := suffixarray.New(bin)

	subsCounts := map[string]struct{}{}
	var freq []subFreq

	for l := *minLen; l <= *maxLen; l++ {
		log.Printf("counting strings of length %d (up to %d)", l, *maxLen)
		for i := 0; i < len(bin)-l; i++ {
			subs := bin[i : i+l]
			subss := string(subs)
			if _, alreadyCounted := subsCounts[subss]; alreadyCounted {
				continue
			}
			offsets := sa.Lookup(subs, -1)
			subsCounts[subss] = struct{}{}

			if len(offsets) >= *minFreq {
				freq = append(freq, subFreq{Sub: hex.EncodeToString(subs), Freq: len(offsets)})
			}
		}
	}
	log.Printf("sorting results")
	sort.Slice(freq, func(i, j int) bool {
		if freq[i].Freq == freq[j].Freq {
			return len(freq[i].Sub) > len(freq[j].Sub)
		}
		return freq[i].Freq > freq[j].Freq
	})

	log.Printf("encoding results to json")
	if err := json.NewEncoder(os.Stdout).Encode(freq); err != nil {
		log.Fatal(err)
	}
	log.Printf("all done")
}

type subFreq struct {
	Sub  string `json:"sub"`
	Freq int    `json:"freq"`
}
