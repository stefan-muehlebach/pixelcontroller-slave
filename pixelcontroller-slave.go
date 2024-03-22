package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"

    "github.com/stefan-muehlebach/ledgrid"
)

type colorType int

const (
    red colorType = iota
    green
    blue
)

const (
    defPort        = 5333
    defBaud        = 2_000_000
    defGammaString = "3.0,3.0,3.0"
)

func main() {
    var port uint
    var baud int
    var gammaString string
    var gammaValues [3]float64

    var spiDevFile string = "/dev/spidev0.0"

    var pixelServer *ledgrid.PixelServer

    // Verarbeite als erstes die Kommandozeilen-Optionen
    //
    flag.UintVar(&port, "port", defPort, "UDP port")
    flag.IntVar(&baud, "baud", defBaud, "SPI baudrate in Hz")
    flag.StringVar(&gammaString, "gamma", defGammaString, "Gamma values")
    flag.Parse()

    for i, str := range strings.Split(gammaString, ",") {
        val, err := strconv.ParseFloat(str, 64)
        if err != nil {
            log.Fatalf("Wrong format: %s", str)
        }
        gammaValues[i] = val
    }

    pixelServer = ledgrid.NewPixelServer(port, spiDevFile, baud)
    pixelServer.SetGamma(gammaValues[red], gammaValues[green], gammaValues[blue])

    // Damit der Daemon kontrolliert beendet werden kann, installieren wir
    // einen Handler fuer das INT-Signal, welches bspw. durch Ctrl-C erzeugt
    // wird oder auch von systemd beim Stoppen eines Services verwendet wird.
    // Das USR1-Signal wird dafuer verwendet, zu Kontrollzwecken die aktuellen
    // Gamma-Werte auszugeben.
    //
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGUSR1)
        for {
            sig := <-sigChan
            switch sig {
            case os.Interrupt:
                pixelServer.Close()
                return
            case syscall.SIGUSR1:
                gRed, gGreen, gBlue := pixelServer.Gamma()
                log.Printf("Current gamma values for red, green, blue: %f, %f, %f\n", gRed, gGreen, gBlue)
            }
        }
    }()

    pixelServer.Handle()
}
