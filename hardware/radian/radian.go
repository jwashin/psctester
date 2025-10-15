package main

// Converted from python radian.py -> Go
// Single-file implementation. Requires github.com/tarm/serial
// go get github.com/tarm/serial

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/tarm/serial"
)

const (
	Version = "2.02"
)

var Qlogger = log.Default()

type Packet struct {
	Start    byte
	Type     byte
	Length   uint16
	Data     []byte
	Checksum uint16

	// internal
	cBuf []byte
	rBuf []byte
}

func checksum(buf []byte) uint16 {
	var tot uint32
	for _, b := range buf {
		tot += uint32(b)
	}
	return uint16(tot & 0xffff)
}

func (p *Packet) FormatBuf() {
	buf := []byte{p.Start, p.Type}
	lenb := make([]byte, 2)
	binary.BigEndian.PutUint16(lenb, p.Length)
	buf = append(buf, lenb...)
	if p.Length > 0 && len(p.Data) > 0 {
		buf = append(buf, p.Data...)
	}
	cs := checksum(buf)
	csB := make([]byte, 2)
	binary.BigEndian.PutUint16(csB, cs)
	buf = append(buf, csB...)
	p.cBuf = buf
}

func (p *Packet) ProcessRbuf() bool {
	if len(p.rBuf) < 2 {
		return false
	}
	p.Checksum = binary.BigEndian.Uint16(p.rBuf[len(p.rBuf)-2:])
	if checksum(p.rBuf[:len(p.rBuf)-2]) != p.Checksum {
		return false
	}
	if len(p.rBuf) < 4 {
		return false
	}
	p.Start = p.rBuf[0]
	p.Type = p.rBuf[1]
	p.Length = binary.BigEndian.Uint16(p.rBuf[2:4])
	if p.Start == 0xA9 { // NAK
		return false
	}
	if int(p.Length) > 0 && len(p.rBuf) >= 4+int(p.Length) {
		p.Data = make([]byte, p.Length)
		copy(p.Data, p.rBuf[4:4+int(p.Length)])
	} else {
		p.Data = nil
	}
	return true
}

type Radian struct {
	port       *serial.Port
	outPacket  *Packet
	inPacket   *Packet
	ndPtr      int
	model      string
	serialNo   string
	name       string
	version    string
	calDate    string
	volts      float64
	amps       float64
	watts      float64
	voltAmps   float64
	VARs       float64
	frequency  float64
	degPhase   float64
	powerFact  float64
	deltaPhase float64
	wattHours  float64
	// ... other accumulated fields
}

var pNameFromCode = map[byte]struct {
	Name  string
	Flags byte
}{
	0x00: {"NOP", 0},
	0x02: {"Identification", 1},
	0x03: {"Reset RD", 0},
	0x04: {"Read Calibration", 1},
	0x07: {"Reset Metrics", 0},
	0x08: {"Start Accumulating Metrics", 0},
	0x09: {"Stop Accumulating Metrics", 0},
	0x0a: {"Start A Timed Accumulating Test", 0},
	0x0b: {"Lock/Unlock Relay Ranges", 0},
	0x0c: {"Trigger Waveform", 0},
	0x0d: {"Read Instantaneous Metrics RD-2x format", 1},
	0x0e: {"Accumulated Waveform Data Read", 1},
	0x0f: {"Harmonic Data Read", 1},
	0x16: {"Read Accumulated Metrics RD-2x format", 1},
	0x1b: {"Auto-Calibrate", 0},
	0x1d: {"BNC Control", 0},
	0x20: {"System Status", 1},
	0x21: {"Minimum Metrics Data Read RD-2x format", 1},
	0x23: {"Maximum Metrics Data Read RD-2x format", 1},
	0x28: {"Trigger Harmonic Analysis", 1},
	0x2c: {"Mode Change", 0},
	0x2e: {"Read Instantaneous Metrics RD-3x format", 1},
	0x2f: {"Read Accumulated Metrics RD-3x format", 1},
	0x30: {"Minimum Metrics Data Read RD-3x format", 1},
	0x31: {"Maximum Metrics Data Read RD-3x format", 1},
	0x32: {"Pulse Output Constant Change", 0},
	0x34: {"Standard Test", 1},
	0x39: {"Meter Test", 1},
}

var pCodeFromName = func() map[string]byte {
	m := make(map[string]byte)
	for k, v := range pNameFromCode {
		m[v.Name] = k
	}
	return m
}()

func NewRadian(tty string) (*Radian, error) {
	c := &serial.Config{
		Name:        tty,
		Baud:        57600,
		ReadTimeout: time.Millisecond * 500,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	r := &Radian{
		port:      s,
		outPacket: &Packet{Start: 0xA6},
		inPacket:  &Packet{},
	}
	// initial commands as original
	_ = r.SendCmd("Identification", nil)
	_ = r.SendCmd("Read Calibration", []byte{0x00, 0x08, 0xA7, 0x00, 0x4F})
	return r, nil
}

func (r *Radian) nextString() string {
	if r.inPacket == nil || r.inPacket.Data == nil {
		return ""
	}
	data := r.inPacket.Data
	i := r.ndPtr
	var b byte
	var sb strings.Builder
	for ; i < len(data); i++ {
		b = data[i]
		if b == ',' {
			break
		}
		if b != 0x00 {
			sb.WriteByte(b)
		}
	}
	r.ndPtr = i + 1
	return sb.String()
}

func tiFloat(pass uint32) float64 {
	expon := int((pass >> 24) & 0xff)
	if expon == 0x80 || expon == 0x81 {
		return 0.0
	}
	if expon > 127 {
		expon = expon - 256
	}
	mant := int64(pass & 0x7fffff)
	if (pass & 0x800000) != 0 {
		mant = mant - (1 << 24)
	} else {
		mant = mant | (1 << 23)
	}
	return (float64(mant) / float64(1<<23)) * math.Pow(2, float64(expon))
}

func (r *Radian) SendCmd(msgType string, msgData []byte) error {
	code, ok := pCodeFromName[msgType]
	if !ok {
		return fmt.Errorf("unknown command name: %s", msgType)
	}
	r.outPacket.Type = code
	if msgData == nil {
		r.outPacket.Length = 0
		r.outPacket.Data = nil
	} else {
		r.outPacket.Length = uint16(len(msgData))
		r.outPacket.Data = msgData
	}
	r.outPacket.FormatBuf()

	try := 3
	var lastErr error
	for try > 0 {
		_, err := r.port.Write(r.outPacket.cBuf)
		if err != nil {
			lastErr = err
			try--
			continue
		}
		// read available bytes (up to 400)
		buf := make([]byte, 400)
		n, err := r.port.Read(buf)
		if err != nil {
			lastErr = err
			try--
			continue
		}
		r.inPacket.rBuf = make([]byte, n)
		copy(r.inPacket.rBuf, buf[:n])
		if r.inPacket.ProcessRbuf() {
			break
		}
		lastErr = fmt.Errorf("invalid response (attempts left %d)", try-1)
		try--
	}
	if try == 0 && lastErr != nil {
		return lastErr
	}

	// Handle response types
	switch r.inPacket.Type {
	case 0x02:
		r.ndPtr = 0
		r.model = r.nextString()
		r.serialNo = r.nextString()
		r.name = r.nextString()
		r.version = r.nextString()
		Qlogger.Printf("Radian version: %s,%s,%s,%s", r.model, r.serialNo, r.name, r.version)
	case 0x04:
		// data layout original code: date bytes at offsets 0:4 year? and month/day at 4:6 etc.
		if len(r.inPacket.Data) >= 8 {
			year := string(r.inPacket.Data[0:4])
			month := string(r.inPacket.Data[4:6])
			day := string(r.inPacket.Data[6:8])
			r.calDate = month + "/" + day + "/" + year
			Qlogger.Printf("Radian cal date: %s", r.calDate)
		}
	case 0x0d:
		if len(r.inPacket.Data) >= 9*4 {
			h := make([]uint32, 9)
			for i := 0; i < 9; i++ {
				h[i] = binary.BigEndian.Uint32(r.inPacket.Data[i*4 : i*4+4])
			}
			r.volts = tiFloat(h[0])
			r.amps = tiFloat(h[1])
			r.watts = tiFloat(h[2])
			r.voltAmps = tiFloat(h[3])
			r.VARs = tiFloat(h[4])
			r.frequency = tiFloat(h[5])
			r.degPhase = tiFloat(h[6])
			r.powerFact = tiFloat(h[7])
			r.deltaPhase = tiFloat(h[8])
			Qlogger.Printf("Radian Inst. values: %.4gV %.4gA %.4gW %.4gVA %.4gVAR %.4gHz %.4gdeg %.4gPF %.4gdeg",
				r.volts, r.amps, r.watts, r.voltAmps, r.VARs, r.frequency, r.degPhase, r.powerFact, r.deltaPhase)
		}
	case 0x16:
		if len(r.inPacket.Data) >= 8*4 {
			h := make([]uint32, 8)
			for i := 0; i < 8; i++ {
				h[i] = binary.BigEndian.Uint32(r.inPacket.Data[i*4 : i*4+4])
			}
			r.wattHours = tiFloat(h[0])
			// other accumulated fields could be assigned similarly...
			Qlogger.Printf("Radian Acc. values: %.4gWH ...", r.wattHours)
		}
	default:
		// if flag bit0 set => needs unpacking; handled above for known types
		if info, ok := pNameFromCode[r.inPacket.Type]; ok {
			if info.Flags&0x01 != 0 {
				// receive packet needs unpacking - but unknown handler
			}
		}
	}
	return nil
}

func (r *Radian) UpdateReadings() error {
	return r.SendCmd("Read Instantaneous Metrics RD-2x format", []byte{0x00, 0x24, 0x00, 0x00, 0x00, 0x14, 0xff, 0xfd})
}

func (r *Radian) SetupForGate() error {
	if err := r.SendCmd("Stop Accumulating Metrics", nil); err != nil {
		return err
	}
	if err := r.SendCmd("BNC Control", []byte{0x01, 0x01}); err != nil {
		return err
	}
	if err := r.SendCmd("Start Accumulating Metrics", []byte{0x06, 0x00, 0x01}); err != nil {
		return err
	}
	return nil
}

func (r *Radian) StartAccumulating() error {
	return r.SendCmd("Start Accumulating Metrics", []byte{0x01, 0x00, 0x00})
}

func (r *Radian) ReadAccumulated() error {
	return r.SendCmd("Read Accumulated Metrics RD-2x format", []byte{0x00, 0x20, 0x00, 0x00, 0x00, 0x04})
}

func main() {
	fmt.Printf("Radian Go driver version %s\n", Version)
	r, err := NewRadian("/dev/tty_radian")
	if err != nil {
		log.Fatalf("failed to open radian: %v", err)
	}
	// short wait for device
	time.Sleep(100 * time.Millisecond)

	fmt.Println(r.model, r.serialNo, r.name, r.version)
	fmt.Println("Last calibrated:", r.calDate)

	if err := r.UpdateReadings(); err != nil {
		log.Printf("update readings error: %v", err)
	} else {
		fmt.Printf("Volts = %8.3f, Amps = %8.3g, Watts = %8.3g, VA = %8.3g, VAR = %8.3g, Frequency = %8.3f\nDegrees Phase = %8.3f, PF = %8.3f, Delta Phase = %8.3g\n",
			r.volts, r.amps, r.watts, r.voltAmps, r.VARs, r.frequency, r.degPhase, r.powerFact, r.deltaPhase)
	}

	_ = r.SendCmd("Read Calibration", []byte{0x00, 0x08, 0xA7, 0x00, 0x4F})
	fmt.Println(r.calDate)

	_ = r.SendCmd("Accumulated Waveform Data Read", []byte{0x00, 0x08, 0x00, 0x00, 0x00, 0x04})

	_ = r.SetupForGate()

	// simple REPL
	for {
		var input string
		fmt.Print("Enter command (r = reset Radian, q = quit, anything else = display KWH)--->")
		_, err := fmt.Scanln(&input)
		if err != nil {
			// ignore input errors (e.g., EOF)
			break
		}
		switch input {
		case "q":
			return
		case "r":
			_ = r.SetupForGate()
		default:
			_ = r.ReadAccumulated()
			fmt.Printf("KWH (watt_hours) = %v\n", r.wattHours)
		}
	}
}
