module pixelcontroller-slave

go 1.22.0

replace github.com/stefan-muehlebach/ledgrid => ../ledGrid

require (
	github.com/stefan-muehlebach/ledgrid v0.0.0-00010101000000-000000000000
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.2
)
