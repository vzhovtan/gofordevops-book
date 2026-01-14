module push

replace model => /home/vzhovtan/gosrc/gofordevops-book/chapter11/model

replace render => /home/vzhovtan/gosrc/gofordevops-book/chapter11/render

go 1.24.0

toolchain go1.24.11

require (
	golang.org/x/crypto v0.47.0
	model v0.0.0-00010101000000-000000000000
	render v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.40.0 // indirect
