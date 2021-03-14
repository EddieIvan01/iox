module iox

go 1.16

require (
	github.com/klauspost/compress v1.11.12
	github.com/klauspost/reedsolomon v1.9.12 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b // indirect
	github.com/tjfoc/gmsm v1.4.0 // indirect
	github.com/xtaci/kcp-go v5.4.20+incompatible
	github.com/xtaci/lossyconn v0.0.0-20200209145036-adba10fffc37 // indirect
	github.com/xtaci/smux v1.5.15
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210313202042-bd2e13477e9c
)

replace github.com/xtaci/smux v1.5.15 => github.com/eddieivan01/smux v1.5.15
