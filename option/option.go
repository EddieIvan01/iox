package option

const (
	TCP_BUFFER_SIZE = 0x4000

	CONN_BUFFER_SIZE = 0x20

	// If buffer size is not large enough,
	// UDPConn.Read will drop the packet data
	// exceeds buffer size, it's not like
	// stream protocol, so I need to allocate
	// a MAX_SIZE buffer
	UDP_PACKET_MAX_SIZE = 0x10000

	UDP_PACKET_CHANNEL_SIZE = 0x400

	MAX_UDP_FWD_WORKER = 0x10
)

var (
	TIMEOUT = 5000

	PROTOCOL = "TCP"

	// enable log output
	VERBOSE = false

	// logic optimization, changed in v0.1.1
	FORWARD_WITHOUT_DEC = false
)
