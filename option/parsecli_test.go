package option

import "testing"

func TestParseCli(t *testing.T) {
	var mode string
	var submode int
	var local, remote []string
	var lenc, renc []bool
	var err error

	mode, submode, local, remote, lenc, renc, err = ParseCli([]string{"fwd", "-l", "9999", "-r", "1.1.1.1:8888", "-k", "0001", "-v"})
	if mode != "fwd" || submode != SUBMODE_L2R || lenc[0] || renc[0] || local[0] != ":9999" || remote[0] != "1.1.1.1:8888" || err != nil {
		t.Error("Error case 1")
	}

	mode, submode, local, remote, lenc, renc, err = ParseCli([]string{"fwd", "-l", "9999", "-l", "*8888", "-k", "0001", "-v"})
	if mode != "fwd" || submode != SUBMODE_L2L || lenc[0] || !lenc[1] || local[0] != ":9999" || local[1] != ":8888" || err != nil {
		t.Error("Error case 2")
	}

	mode, submode, local, remote, lenc, renc, err = ParseCli([]string{"fwd", "-r", "*1.1.1.1:9999", "-r", "*1.1.1.1:8888", "-k", "0001", "-v"})
	if mode != "fwd" || submode != SUBMODE_R2R || !renc[0] || !renc[1] || remote[0] != "1.1.1.1:9999" || remote[1] != "1.1.1.1:8888" || err != nil {
		t.Error(mode, submode, local, remote, lenc, renc, err, "Error case 3")
	}

	mode, submode, local, remote, lenc, renc, err = ParseCli([]string{"proxy", "-r", "*1.1.1.1:9999", "-r", "*1.1.1.1:8888", "-k", "0001", "-v"})
	if mode != "proxy" || err != errUnrecognizedSubMode {
		t.Error("Error case 4")
	}

	mode, submode, local, remote, lenc, renc, err = ParseCli([]string{"fwd", "-l", ":9999", "-r", "1.1.1.1:8888", "-k", "0001", "-h"})
	if mode != "fwd" || err != PrintUsage {
		t.Error("Error case 5")
	}
}
