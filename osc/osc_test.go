package osc

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
	"time"
)

// -- Fake PacketConn to capture WriteTo calls --------------------

type fakeConn struct {
	lastData []byte
	lastAddr net.Addr
}

func (f *fakeConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	f.lastData = append([]byte(nil), b...)
	f.lastAddr = addr
	return len(b), nil
}
func (f *fakeConn) ReadFrom(p []byte) (int, net.Addr, error) { return 0, nil, nil }
func (f *fakeConn) Close() error                             { return nil }
func (f *fakeConn) LocalAddr() net.Addr                      { return &net.UDPAddr{} }
func (f *fakeConn) SetDeadline(t time.Time) error            { return nil }
func (f *fakeConn) SetReadDeadline(t time.Time) error        { return nil }
func (f *fakeConn) SetWriteDeadline(t time.Time) error       { return nil }

// -- Tests for the low-level builders ---------------------------------

func Test_PadString(t *testing.T) {
	testCases := []struct {
		input string
	}{
		{"action"},
		{"4077"},
		{"_S&M_INS_MARKER_PLAY"},
		{"streamdeck"},
		{"osc"},
		{"a"},
		{"abc"},
		{"1234567890"},
	}

	for _, tc := range testCases {
		result := PadString(tc.input)
		// Check if the length of the byte slice is divisible by 4
		if len(result)%4 != 0 {
			t.Errorf("Length of padded string for input %q is %d; expected to be divisible by 4", tc.input, len(result))
		}
	}
}

func Test_CreateOSCIntPacket(t *testing.T) {
	addr := "/action"
	value := 0x12345678
	packet := CreateOSCIntPacket(addr, value)

	// 1) must be 4-byte aligned
	if len(packet)%4 != 0 {
		t.Errorf("int packet length = %d; want multiple of 4", len(packet))
	}

	// 2) prefix bytes match PadString(addr)
	wantAddr := PadString(addr)
	if !bytes.Equal(packet[:len(wantAddr)], wantAddr) {
		t.Errorf("prefix = %v; want %v", packet[:len(wantAddr)], wantAddr)
	}

	// 3) next bytes = PadString(",i")
	iTag := PadString(",i")
	if !bytes.Equal(packet[len(wantAddr):len(wantAddr)+len(iTag)], iTag) {
		t.Errorf("type-tag = %v; want %v", packet[len(wantAddr):len(wantAddr)+len(iTag)], iTag)
	}

	// 4) final 4 bytes big-endian(value)
	argOffset := len(wantAddr) + len(iTag)
	got := packet[argOffset : argOffset+4]
	want := make([]byte, 4)
	binary.BigEndian.PutUint32(want, uint32(value))
	if !bytes.Equal(got, want) {
		t.Errorf("integer bytes = %v; want %v", got, want)
	}
}

func Test_CreateOSCStringPacket(t *testing.T) {
	addr := "/action"
	val := "_S&M_INS_MARKER_PLAY"
	packet := CreateOSCStringPacket(addr, val)

	if len(packet)%4 != 0 {
		t.Errorf("string packet length = %d; want multiple of 4", len(packet))
	}

	// should start with PadString(addr)+PadString(",s")
	prefLen := len(PadString(addr)) + len(PadString(",s"))
	if !bytes.Equal(packet[:prefLen], append(PadString(addr), PadString(",s")...)) {
		t.Error("header mismatch in CreateOSCStringPacket")
	}

	// should end with PadString(val)
	if !bytes.Equal(packet[len(packet)-len(PadString(val)):], PadString(val)) {
		t.Errorf("trailer = %v; want %v", packet[len(packet)-len(PadString(val)):], PadString(val))
	}
}

func Test_CreateOSCPathPacket(t *testing.T) {
	full := "/midiaction/_FX_ABC123"
	got := CreateOSCPathPacket(full)
	want := PadString(full)
	if !bytes.Equal(got, want) {
		t.Errorf("path packet = %v; want %v", got, want)
	}
}

// -- Tests for the dispatcher SendOSC -----------------------------

func Test_SendOSC_Integer(t *testing.T) {
	fc := &fakeConn{}
	// pick a non‐zero port so string formatting is exercised
	SendOSC("127.0.0.1", 54321, "/action", "123", fc)

	// remote address must resolve to 127.0.0.1:54321
	if fc.lastAddr.String() != "127.0.0.1:54321" {
		t.Errorf("addr = %q; want %q", fc.lastAddr, "127.0.0.1:54321")
	}
	// lastData must equal CreateOSCIntPacket("/action", 123)
	want := CreateOSCIntPacket("/action", 123)
	if !reflect.DeepEqual(fc.lastData, want) {
		t.Error("Sent packet does not match expected int packet")
	}
}

func Test_SendOSC_StringMain(t *testing.T) {
	fc := &fakeConn{}
	SendOSC("localhost", 11111, "/action", "FOO_BAR", fc)

	// port and host formatting
	if fc.lastAddr.String() != "127.0.0.1:11111" && fc.lastAddr.String() != "localhost:11111" {
		t.Errorf("addr = %q; want localhost:11111", fc.lastAddr)
	}
	want := CreateOSCStringPacket("/action", "FOO_BAR")
	if !bytes.Equal(fc.lastData, want) {
		t.Error("Sent packet does not match expected string-arg packet")
	}
}

func Test_SendOSC_StringMidi(t *testing.T) {
	fc := &fakeConn{}
	SendOSC("1.2.3.4", 9999, "/midiaction", "_S&M_MARK", fc)

	if fc.lastAddr.String() != "1.2.3.4:9999" {
		t.Errorf("addr = %q; want 1.2.3.4:9999", fc.lastAddr)
	}
	// for MIDI non-numeric, should be path-only
	want := CreateOSCPathPacket("/midiaction/_S&M_MARK")
	if !bytes.Equal(fc.lastData, want) {
		t.Errorf("Sent packet = %v; want path-only %v", fc.lastData, want)
	}
}
