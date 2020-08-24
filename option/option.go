package option

const (
	TCP_BUFFER_SIZE = 0x8000

	// UDP protocol's max capacity
	UDP_PACKET_MAX_SIZE = 0xFFFF - 28

	UDP_PACKET_CHANNEL_SIZE = 0x400

	HEARTBEAT_FREQUENCY = 30

	CONNECTING_RETRY_DURATION = 1500
)

var (
	TIMEOUT = 5000

	PROTOCOL = "TCP"

	// enable log output
	VERBOSE = false

	// logic optimization, changed in v0.1.1
	FORWARD_WITHOUT_DEC = false
)
