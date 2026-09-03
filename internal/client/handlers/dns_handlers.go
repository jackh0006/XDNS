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
	RegisterHandler(Enums.PACKET_DNS_QUERY_REQ_ACK, handleDNSQueryAck)
	RegisterHandler(Enums.PACKET_DNS_QUERY_RES, handleDNSQueryRes)
}

func handleDNSQueryAck(c ClientContext, packet VpnProto.Packet, addr *net.UDPAddr) error {
	return c.HandleDNSQueryAck(packet)
}

func handleDNSQueryRes(c ClientContext, packet VpnProto.Packet, addr *net.UDPAddr) error {
	return c.HandleDNSQueryRes(packet)
}
