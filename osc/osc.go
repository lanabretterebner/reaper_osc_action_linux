package osc

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
)

// padString pads a string to a multiple of 4 bytes by adding null bytes (0x00).
func PadString(input string) []byte {
	// Start by appending a null terminator to the input string
	strWithNull := input + "\x00"

	length := len(strWithNull)

	padding := (4 - (length % 4)) % 4

	// Append the necessary number of null bytes (0-3)
	paddedString := strWithNull + string(bytes.Repeat([]byte{'\x00'}, padding))

	result := []byte(paddedString)
	return result
}

func CreateOSCIntPacket(address string, value int) []byte {
	var buf bytes.Buffer
	buf.Write(PadString(address)) // "/action" or "/midiaction"
	buf.Write(PadString(",i"))    // int type tag
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(value))
	buf.Write(b[:])
	return buf.Bytes()
}

// Create a “string‐arg” OSC packet: e.g. /action [s] "_S&M_…"
func CreateOSCStringPacket(address, val string) []byte {
	var buf bytes.Buffer
	buf.Write(PadString(address)) // "/action"
	buf.Write(PadString(",s"))    // string type tag
	buf.Write(PadString(val))     // the actual command‐ID
	return buf.Bytes()
}

// Create a path‐only OSC packet (no type tags): e.g. "/midiaction/_S&M_…"
func CreateOSCPathPacket(fullPath string) []byte {
	// just send the padded address – Reaper will see it as “no args” ([])
	return PadString(fullPath)
}

func SendOSC(host string, port int, prefix, commandID string, udp_client net.PacketConn) {
	var packet []byte

	// 1) main‐window
	if prefix == "/action" {
		if n, err := strconv.Atoi(commandID); err == nil {
			packet = CreateOSCIntPacket(prefix, n)
		} else {
			packet = CreateOSCStringPacket(prefix, commandID)
		}

		// 2) MIDI‐editor
	} else if prefix == "/midiaction" {
		if n, err := strconv.Atoi(commandID); err == nil {
			packet = CreateOSCIntPacket(prefix, n)
		} else {
			// MIDI‐editor string commands must be path‐only
			packet = CreateOSCPathPacket(prefix + "/" + commandID)
		}

		// 3) fallback (if you ever add other prefixes)
	} else {
		if n, err := strconv.Atoi(commandID); err == nil {
			packet = CreateOSCIntPacket(prefix, n)
		} else {
			packet = CreateOSCPathPacket(prefix + "/" + commandID)
		}
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	remote, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("OSC: bad address %s\n", addr)
		return
	}
	if _, err := udp_client.WriteTo(packet, remote); err != nil {
		log.Printf("OSC write error: %v\n", err)
	}
}
