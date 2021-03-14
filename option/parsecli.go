package option

import (
	"encoding/hex"
	"errors"
	"iox/crypto"
	"strconv"
)

var (
	errUnrecognizedMode        = errors.New("Unrecognized mode. Must choose a working mode in [fwd/proxy]")
	errHexDecodeError          = errors.New("Key must be a hexadecimal string")
	PrintUsage                 = errors.New("")
	errUnrecognizedSubMode     = errors.New("Malformed args. Incorrect number of socket descriptors")
	errNoSecretKey             = errors.New("Encryption enabled, must specify a key by `-k` param")
	errNotANumber              = errors.New("Timeout param must be a number")
	errUDPProxyMode            = errors.New("Unsupported UDP proxy mode")
	errFwdUnreliableBtReliable = errors.New("Can't forward between unreliable and reliable protocols")
)

const (
	SUBMODE_L2L byte = iota
	SUBMODE_R2R
	SUBMODE_L2R

	SUBMODE_LP
	SUBMODE_RP
	SUBMODE_RPL2L
)

func ParseCli(args []string) (
	mode string,
	submode byte,
	descs []*SocketDesc,
	err error) {

	if len(args) == 0 {
		err = PrintUsage
		return
	}

	mode = args[0]

	switch mode {
	case "fwd", "proxy":
	case "-h", "--help":
		err = PrintUsage
		return
	default:
		err = errUnrecognizedMode
		return
	}

	args = args[1:]
	ptr := 0

	for {
		if ptr == len(args) {
			break
		}

		if args[ptr][0] != '-' {
			var desc *SocketDesc
			desc, err = NewSocketDesc(args[ptr])
			if err != nil {
				return
			}
			descs = append(descs, desc)

			ptr++
			continue
		}

		switch args[ptr] {
		case "-k", "--key":
			var key []byte
			key, err = hex.DecodeString(args[ptr+1])
			if err != nil {
				err = errHexDecodeError
				return
			}
			crypto.ExpandKey(key)
			ptr++

		case "-t", "--timeout":
			TIMEOUT, err = strconv.Atoi(args[ptr+1])
			if err != nil {
				err = errNotANumber
				return
			}
			ptr++
		case "-v", "--verbose":
			VERBOSE = true
		case "-h", "--help":
			err = PrintUsage
			return
		}

		ptr++
	}

	if mode == "fwd" {
		switch {
		case len(descs) != 2:
			err = errUnrecognizedSubMode
			return
		case descs[0].IsListener && descs[1].IsListener:
			submode = SUBMODE_L2L
		case !descs[0].IsListener && !descs[1].IsListener:
			submode = SUBMODE_R2R
		default:
			submode = SUBMODE_L2R
		}
	} else {
		switch {
		case len(descs) == 1 && descs[0].IsProxyProto():
			submode = SUBMODE_LP
		case len(descs) == 1 && !descs[0].IsProxyProto():
			submode = SUBMODE_RP
		case len(descs) == 2 &&
			((descs[0].IsListener && descs[1].IsProxyProto()) ||
				(descs[1].IsListener && descs[0].IsProxyProto())):
			submode = SUBMODE_RPL2L
		default:
			err = errUnrecognizedSubMode
			return
		}
	}

	if crypto.SECRET_KEY == nil {
		for i := range descs {
			if descs[i].Secret {
				err = errNoSecretKey
				return
			}
		}
	}

	if mode == "fwd" && ((descs[0].IsProtoReliable() && !descs[1].IsProtoReliable()) ||
		(!descs[0].IsProtoReliable() && descs[1].IsProtoReliable())) {
		err = errFwdUnreliableBtReliable
		return
	}

	if mode == "fwd" && len(descs) == 2 {
		if descs[0].Secret && descs[1].Secret {
			FORWARD_WITHOUT_DEC = true
		}

		if descs[0].Compress && descs[1].Compress {
			FORWARD_WITHOUT_COMPRESS = true
		}
	}

	return
}
