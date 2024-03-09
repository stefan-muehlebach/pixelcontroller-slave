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
    defGammaValues = "3.0,3.0,3.0"
    defBaud        = 2_000_000
    defUseTCP      = false

    bufferSize = 1024
)

func main() {
    var port uint
    var baud int
    var gammaValues string
    var gammaValue [3]float64

    var spiDevFile string = "/dev/spidev0.0"

    var pixelServer *ledgrid.PixelServer

    // Verarbeite als erstes die Kommandozeilen-Optionen
    //
    flag.UintVar(&port, "port", defPort, "UDP port")
    flag.IntVar(&baud, "baud", defBaud, "SPI baudrate in Hz")
    flag.StringVar(&gammaValues, "gamma", defGammaValues, "Gamma values")
    flag.Parse()

    for i, str := range strings.Split(gammaValues, ",") {
        val, err := strconv.ParseFloat(str, 64)
        if err != nil {
            log.Fatalf("Wrong format: %s", str)
        }
        gammaValue[i] = val
    }

    pixelServer = ledgrid.NewPixelServer(port, spiDevFile, baud)
    pixelServer.SetGamma(0, gammaValue[red])
    pixelServer.SetGamma(1, gammaValue[green])
    pixelServer.SetGamma(2, gammaValue[blue])

    // Damit der Daemon kontrolliert beendet werden kann, installieren wir
    // einen Handler fuer das INT-Signal, welches bspw. durch Ctrl-C erzeugt
    // wird oder auch von systemd beim Stoppen eines Services verwendet wird.
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
                gRed   := pixelServer.Gamma(0)
                gGreen := pixelServer.Gamma(1)
                gBlue  := pixelServer.Gamma(2)
                log.Printf("Current gamma values for red, green, blue: %f, %f, %f\n", gRed, gGreen, gBlue)
            }
        }
    }()

    pixelServer.Handle()
}

