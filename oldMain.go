//go:build ignore
// +build ignore

package main

import (
	"errors"
	"flag"
	"log"
	"math"
	"net"
	"net/netip"
	"os"
	"os/signal"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
    "periph.io/x/host/v3/rpi"
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

	var gamma [3][256]byte
	var err error
    var onRaspi bool

	var addrPort netip.AddrPort
	var udpAddr *net.UDPAddr
	var udpConn *net.UDPConn
	var buffer []byte
	var len int

	var spiDevFile string = "/dev/spidev0.0"
	var spiBaud physic.Frequency
	var spiPort spi.PortCloser
	var spiConn spi.Conn

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

	spiBaud = physic.Frequency(baud)

	// Anschliessend wird die Tabelle fuer die Farbwertkorrektur erstellt.
	//
	for color := red; color <= blue; color++ {
		for i := 0; i < 256; i++ {
			gamma[color][i] = byte(255.0 * math.Pow(float64(i)/255.0,
				gammaValue[color]))
		}
	}

	// Dann erstellen wir einen Buffer fuer die via Netzwerk eintreffenden
	// Daten. 1kB sollten aktuell reichen (entspricht rund 340 RGB-Werten).
	//
	buffer = make([]byte, bufferSize)

	// Dann wird der SPI-Bus initialisiert.
	//
	_, err = host.Init()
	if err != nil {
		log.Fatal(err)
	}
    if rpi.Present() {
        onRaspi = true
    }

    if onRaspi {
        	spiPort, err = spireg.Open(spiDevFile)
        	if err != nil {
        		log.Fatal(err)
        	}
        	defer spiPort.Close()
        	spiConn, err = spiPort.Connect(spiBaud*physic.Hertz, spi.Mode0, 8)
        	if err != nil {
        		log.Fatal(err)
        	}
    }

	// Jetzt wird der UDP-Port geoeffnet, resp. eine lesende Verbindung
	// dafuer erstellt.
	//
	addrPort = netip.AddrPortFrom(netip.IPv4Unspecified(), uint16(port))
	if !addrPort.IsValid() {
		log.Fatalf("Invalid address or port")
	}
	udpAddr = net.UDPAddrFromAddrPort(addrPort)
	udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Damit der Daemon kontrolliert beendet werden kann, installieren wir
	// einen Handler fuer das INT-Signal, welches bspw. durch Ctrl-C erzeugt
	// wird oder auch von systemd beim Stoppen eines Services verwendet wird.
	//
	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		udpConn.Close()
	}()

	// Dies schliesslich ist die Schleife, welche das Hauptprogramm ausmacht.
	// Sie laeuft endlos, resp. wird erst beendet, wenn der Signal-Handler
	// (siehe oben) die UDP-Verbindung schliesst und damit die Methode
	// Read() zum Abbruch zwingt.
	//
	for {
		len, err = udpConn.Read(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Fatal(err)
		}
		if len != 300 {
			log.Printf("Only %d bytes received instead of 300.\n", len)
		}
		for i := 0; i < len; i += 3 {
			buffer[i+0] = gamma[red][buffer[i+0]]
			buffer[i+1] = gamma[green][buffer[i+1]]
			buffer[i+2] = gamma[blue][buffer[i+2]]
		}
        if onRaspi {
		    if err = spiConn.Tx(buffer[:len], nil); err != nil {
		    	    log.Printf("Error during communication via SPI: %v\n", err)
		    }
        } else {
            log.Printf("Received %d bytes", len)
        }
	}

	// Vor dem Beenden des Programms werden alle LEDs Schwarz geschaltet
	// damit das Panel dunkel wird.
	//
	for i := range buffer {
		buffer[i] = 0x00
	}
    if onRaspi {
	    if err = spiConn.Tx(buffer, nil); err != nil {
            log.Printf("Error during communication via SPI: %v\n", err)
        }
    } else {
        log.Printf("Turning all LEDs off.")
    }
}
