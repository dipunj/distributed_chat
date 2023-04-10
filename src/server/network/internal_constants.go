package network

const (
	// listen on all interfaces
	DEFAULT_INTERFACE     = "0.0.0.0"
	DEFAULT_PUBLIC_PORT   = "12000"
	DEFAULT_INTERNAL_PORT = "11000"

	PUBLIC_ADDRESS   = DEFAULT_INTERFACE + ":" + DEFAULT_PUBLIC_PORT
	INTERNAL_ADDRESS = DEFAULT_INTERFACE + ":" + DEFAULT_INTERNAL_PORT
)
