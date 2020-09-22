package main

import (
	"fmt"
	"iox/operate"
	"iox/option"
	"os"
)

const VERSION = "0.4"

func Usage() {
	fmt.Printf(
		"iox v%v\n"+
			"    Access intranet easily (https://github.com/eddieivan01/iox)\n\n"+
			"Usage: iox fwd/proxy [-l [*][HOST:]PORT] [-r [*]HOST:PORT] [-k HEX] [-t TIMEOUT] [-u] [-h] [-v]\n\n"+
			"Options:\n"+
			"  -l [*][HOST:]PORT\n"+
			"      address to listen on. `*` means encrypted socket\n"+
			"  -r [*]HOST:PORT\n"+
			"      remote host to connect, HOST can be IP or Domain. `*` means encrypted socket\n"+
			"  -k HEX\n"+
			"      hexadecimal format key, be used to generate Key and IV\n"+
			"  -u\n"+
			"      udp forward mode\n"+
			"  -t TIMEOUT\n"+
			"      set connection timeout(millisecond), default is 5000\n"+
			"  -v\n"+
			"      enable log output\n"+
			"  -h\n"+
			"      print usage then exit\n", VERSION,
	)
}

func main() {
	mode, submode, local, remote, lenc, renc, err := option.ParseCli(os.Args[1:])
	if err != nil {
		if err == option.PrintUsage {
			Usage()
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	switch mode {
	case "fwd":
		switch submode {
		case option.SUBMODE_L2R:
			operate.Local2Remote(local[0], remote[0], lenc[0], renc[0])
		case option.SUBMODE_L2L:
			operate.Local2Local(local[0], local[1], lenc[0], lenc[1])
		case option.SUBMODE_R2R:
			operate.Remote2Remote(remote[0], remote[1], renc[0], renc[1])
		}
	case "proxy":
		switch submode {
		case option.SUBMODE_LP:
			operate.ProxyLocal(local[0], lenc[0])
		case option.SUBMODE_RP:
			operate.ProxyRemote(remote[0], renc[0])
		case option.SUBMODE_RPL2L:
			operate.ProxyRemoteL2L(local[0], local[1], lenc[0], lenc[1])
		}
	}
}
