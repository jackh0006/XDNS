// ==============================================================================
// XDNS
// Author: tajirax
// Github: https://github.com/WhiteDNS/XDNS
// Year: 2026
// ==============================================================================
package handlers

import (
	"net"

	Enums "xdns-go/internal/enums"
	VpnProto "xdns-go/internal/vpnproto"
)

func init() {
	RegisterHandler(Enums.PACKET_MTU_UP_RES, handleMTUResponse)
	RegisterHandler(Enums.PACKET_MTU_DOWN_RES, handleMTUResponse)
}

func handleMTUResponse(c ClientContext, packet VpnProto.Packet, addr *net.UDPAddr) error {
	return c.HandleMTUResponse(packet)
}
