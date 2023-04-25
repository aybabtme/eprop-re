package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

func main() {
	device := flag.String("dev", "/dev/tty.usbserial-A105BLO7", "device to open")
	baud := flag.Int("baud", 115200, "baud rate")
	splitOnHex := flag.String("split-on", "2928", "bytes to split on, in hex")
	flag.Parse()

	splitOn, err := hex.DecodeString(*splitOnHex)
	if err != nil {
		log.Fatal(err)
	}

	mode := &serial.Mode{
		BaudRate: *baud,
	}
	port, err := serial.Open(*device, mode)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		time.Sleep(200 * time.Millisecond)
		writeByte(port, 0x00)
	}
	tryRead(port, []byte("BBIO1"))
	time.Sleep(1000 * time.Millisecond)
	writeByte(port, 0b00000011)
	mustRead(port, []byte("ART1"))

	time.Sleep(1000 * time.Millisecond)
	writeByte(port, 0b00000001)
	mustRead(port, []byte("ART1"))

	time.Sleep(1000 * time.Millisecond)
	writeByte(port, 0b01100111)
	mustRead(port, []byte{0x01})

	time.Sleep(1000 * time.Millisecond)
	writeByte(port, 0b10010000)
	mustRead(port, []byte{0x01})

	go func() {
		scan := bufio.NewScanner(port)
		scan.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.Index(data, splitOn); i >= 0 {
				return i + 1, data[0:i], nil
			}

			return 0, nil, nil
		})
		for scan.Scan() {
			data := scan.Bytes()
			fmt.Fprintf(os.Stdout, `{"t":%q,"v":%q}`+"\n", time.Now().Format(time.RFC3339Nano), hex.EncodeToString(data))
		}

		if err := scan.Err(); err != nil {
			log.Fatalf("scanning: %v", err)
		}
	}()

	time.Sleep(1000 * time.Millisecond)
	writeByte(port, 0b00001111)

	select {}
}

func writeByte(port io.Writer, b byte) {
	log.Printf("writing %x (%b)", b, b)
	_, err := port.Write([]byte{b})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("done")
}

func mustRead(port io.Reader, b []byte) {
	log.Printf("reading %x", b)
	out := make([]byte, len(b))
	if _, err := port.Read(out); err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(out, b) {
		log.Fatalf("unexpected response: %x (%q)", out, string(out))
	}
	log.Printf("done")
}

func tryRead(port io.Reader, b []byte) {
	log.Printf("reading %x", b)
	out := make([]byte, len(b))
	if _, err := port.Read(out); err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(out, b) {
		log.Fatalf("unexpected response: %x", out)
	}
	log.Printf("done")
}
