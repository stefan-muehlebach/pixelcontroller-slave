module pixelcontroller-slave

go 1.22.1

replace github.com/stefan-muehlebach/ledgrid => ../ledgrid

require github.com/stefan-muehlebach/ledgrid v0.0.0-00010101000000-000000000000

require (
	periph.io/x/conn/v3 v3.7.0 // indirect
	periph.io/x/host/v3 v3.8.2 // indirect
)
