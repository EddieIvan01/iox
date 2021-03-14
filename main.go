package main

import (
	"fmt"
	"iox/operate"
	"iox/option"
	"os"
)

const VERSION = "0.5-beta"

func Usage() {
	fmt.Printf(
		"iox v%s\n"+
			"    Access intranet easily (https://github.com/eddieivan01/iox)\n"+
			"    (Protocols: %s)\n\n"+
			"Usage: iox <MODE> [OPTIONS] <SOCKET_DESCRIPTOR> [SOCKET_DESCRIPTOR]\n\n"+
			"Options:\n"+
			"  -k HEX\n"+
			"      hexadecimal format key (required when encryption is enabled)\n"+
			"  -t TIMEOUT\n"+
			"      set connection timeout(millisecond), default is 5000\n"+
			"  -v\n"+
			"      enable log output\n"+
			"  -h\n"+
			"      print usage then exit\n", VERSION, option.SupportedProtocols,
	)
}

func main() {
	mode, submode, descs, err := option.ParseCli(os.Args[1:])
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
			if descs[0].IsListener {
				operate.Local2Remote(descs[0], descs[1])
			} else {
				operate.Local2Remote(descs[1], descs[0])
			}
		case option.SUBMODE_L2L:
			operate.Local2Local(descs[0], descs[1])
		case option.SUBMODE_R2R:
			operate.Remote2Remote(descs[0], descs[1])
		}
	case "proxy":
		switch submode {
		case option.SUBMODE_LP:
			operate.ProxyLocal(descs[0])
		case option.SUBMODE_RP:
			operate.ProxyRemote(descs[0])
		case option.SUBMODE_RPL2L:
			if descs[0].IsProxyProto() {
				operate.ProxyRemoteL2L(descs[1], descs[0])
			} else {
				operate.ProxyRemoteL2L(descs[0], descs[1])
			}
		}
	}
}
