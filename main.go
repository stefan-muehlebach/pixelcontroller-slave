package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/stefan-muehlebach/ledgrid"
)

type colorType int

const (
	red colorType = iota
	green
	blue
)

const (
	defPort       = 5333
	defGammaRed   = 3.0
	defGammaGreen = 3.0
	defGammaBlue  = 3.0
	defBaud       = 2_000_000
	defUseTCP     = false

	bufferSize = 1024
)

func main() {
	var port uint
	var baud int
	var gammaValue [3]float64

	// var gamma [3][256]byte
	// var err error
	// var onRaspi bool

	// var addrPort netip.AddrPort
	// var udpAddr *net.UDPAddr
	// var udpConn *net.UDPConn
	// var buffer []byte
	// var len int

	var spiDevFile string = "/dev/spidev0.0"
	// var spiBaud physic.Frequency
	// var spiPort spi.PortCloser
	// var spiConn spi.Conn

	var pixelServer *ledgrid.PixelServer

	// Verarbeite als erstes die Kommandozeilen-Optionen
	//
	flag.UintVar(&port, "port", defPort, "UDP port")
	flag.IntVar(&baud, "baud", defBaud, "SPI baudrate in Hz")
	flag.Float64Var(&gammaValue[red], "red", defGammaRed,
		"Gamma value for red")
	flag.Float64Var(&gammaValue[green], "green", defGammaGreen,
		"Gamma value for green")
	flag.Float64Var(&gammaValue[blue], "blue", defGammaBlue,
		"Gamma value for blue")
	flag.Parse()

	pixelServer = ledgrid.NewPixelServer(port, spiDevFile, baud)
	pixelServer.SetGamma(0, gammaValue[red])
	pixelServer.SetGamma(1, gammaValue[green])
	pixelServer.SetGamma(2, gammaValue[blue])

	// Damit der Daemon kontrolliert beendet werden kann, installieren wir
	// einen Handler fuer das INT-Signal, welches bspw. durch Ctrl-C erzeugt
	// wird oder auch von systemd beim Stoppen eines Services verwendet wird.
	//
	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		pixelServer.Close()
	}()

	pixelServer.Handle()
}
