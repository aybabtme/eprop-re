package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

func main() {
	device := flag.String("dev", "/dev/tty.usbserial-A105BLO7", "device to open")
	baud := flag.Int("baud", 115200, "baud rate")
	flag.Parse()
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
	time.Sleep(1 * time.Second)
	writeByte(port, 0b00000011)
	mustRead(port, []byte("ART1"))

	time.Sleep(1 * time.Second)
	writeByte(port, 0b00000001)
	mustRead(port, []byte("ART1"))

	time.Sleep(1 * time.Second)
	writeByte(port, 0b01100111)
	mustRead(port, []byte{0x01})

	time.Sleep(1 * time.Second)
	writeByte(port, 0b10010000)
	mustRead(port, []byte{0x01})

	go func() {
		io.Copy(os.Stdout, port)
	}()

	time.Sleep(1 * time.Second)
	writeByte(port, 0b00001111)

	// mustRead(port, []byte("ART1"))

	time.Sleep(20 * time.Second)
	writeByte(port, 0x0)
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
