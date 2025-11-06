// meter.go
//
// Go translation (skeleton) of the provided Python Meter_obj using periph-style interfaces.
// This is a starting point for a Raspberry Pi 3. Replace the TODO periph implementations
// with real periph.io initialisation and opening of UART / I2C devices.
//
// Note: This file is intentionally concise and focuses on structure and core behaviors.
// It uses transport interfaces so you can plug in periph.io-based implementations.

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"
)

// -------- Transport interfaces (implement using periph.io) --------

// SerialPort is a minimal serial interface used by Meter.
// Implement this using periph.io UART driver.
// NOTE: add the following import to the file's import block:
//   "periph.io/x/periph/conn/uart"
//
// Adapter that wraps a periph.io UART port and implements the SerialPort
// interface expected by the rest of the code.
// Use periph.io serial (conn/serial) instead of uart.
// Replace the existing import of "periph.io/x/periph/conn/v3/uart" with:
//   "periph.io/x/periph/conn/v3/serial"

type SerialPort interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	SetBaudrate(int) error
	FlushInput() error
	SetReadTimeout(time.Duration)
	Close() error
}

// Adapter that wraps a periph.io serial port and implements SerialPort.
// Note: change the import at top to use "periph.io/x/periph/conn/v3/serial"
// and ensure the PeriphSerialPort below uses serial.PortCloser.
type PeriphSerialPort struct {
	port        io.ReadWriteCloser
	readTimeout time.Duration
	mu          sync.Mutex
}

// NewPeriphSerialPort returns a SerialPort backed by a periph UART port.
// The caller is responsible for opening the periph UART port (host init, etc)
func NewPeriphSerialPort(p io.ReadWriteCloser) SerialPort {
	return &PeriphSerialPort{
		port:        p,
		readTimeout: 200 * time.Millisecond,
	}
}

func (p *PeriphSerialPort) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.port.Write(b)
}

func (p *PeriphSerialPort) Read(b []byte) (int, error) {
	// Perform a blocking read in a goroutine and enforce read timeout if set.
	type res struct {
		n   int
		err error
	}
	ch := make(chan res, 1)

	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		n, err := p.port.Read(b)
		ch <- res{n: n, err: err}
	}()

	if p.readTimeout <= 0 {
		r := <-ch
		return r.n, r.err
	}
	select {
	case r := <-ch:
		return r.n, r.err
	case <-time.After(p.readTimeout):
		// Timeout - return zero bytes and a timeout error.
		return 0, fmt.Errorf("serial read timeout")
	}
}

func (p *PeriphSerialPort) SetBaudrate(baud int) error {
	// Try to call an underlying SetBaudRate method if provided by the port.
	if s, ok := p.port.(interface{ SetBaudRate(int) error }); ok {
		return s.SetBaudRate(baud)
	}
	// Some periph UART implementations expose SetSpeed with physic.Frequency,
	// try to detect that as well.
	if s2, ok := p.port.(interface{ SetSpeed(uint32) error }); ok {
		return s2.SetSpeed(uint32(baud))
	}
	return fmt.Errorf("underlying port does not support changing baud rate")
}

func (p *PeriphSerialPort) FlushInput() error {
	// Read until no more data is immediately available (drain input).
	buf := make([]byte, 256)
	// Temporarily increase timeout to drain quickly without blocking long.
	oldTimeout := p.readTimeout
	p.readTimeout = 50 * time.Millisecond
	defer func() { p.readTimeout = oldTimeout }()

	for {
		n, err := p.Read(buf)
		if err != nil {
			// treat timeout (no more data) as success
			if strings.Contains(err.Error(), "timeout") {
				return nil
			}
			// other read errors: continue or return
			return nil
		}
		if n == 0 {
			return nil
		}
		// loop to attempt to drain remaining data
	}
}

func (p *PeriphSerialPort) SetReadTimeout(d time.Duration) {
	p.readTimeout = d
}

func (p *PeriphSerialPort) Close() error {
	return p.port.Close()
}

// type SerialPort interface {
// 	Write([]byte) (int, error)
// 	Read([]byte) (int, error)
// 	SetBaudrate(int) error
// 	FlushInput() error
// 	SetReadTimeout(time.Duration)
// 	Close() error
// }

// I2CBus is a minimal I2C interface used by PulseTest.
// Implement this using periph.io I2C driver.

// Add "periph.io/x/periph/conn/i2c" to the file's import block.

type PeriphI2CBus struct {
	bus i2c.BusCloser
	mu  sync.Mutex
}

// NewPeriphI2CBus returns an I2CBus backed by a periph.io i2c.BusCloser.
// The caller is responsible for initializing the periph host and opening the bus.
func NewPeriphI2CBus(b i2c.BusCloser) I2CBus {
	return &PeriphI2CBus{bus: b}
}

func (p *PeriphI2CBus) Tx(addr uint16, w, r []byte) error {
	if p.bus == nil {
		return fmt.Errorf("i2c bus not initialized")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// periph.io i2c.BusCloser implements Tx(addr uint16, w, r []byte) error
	return p.bus.Tx(addr, w, r)
}

func (p *PeriphI2CBus) Close() error {
	if p.bus == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bus.Close()
}

type I2CBus interface {
	Tx(addr uint16, w, r []byte) error
	Close() error
}

// -------- PulseTest (QFAM) minimal stub that uses I2C bus --------

type PulseTest struct {
	bus    I2CBus
	addr   uint16
	meter  *PulseMeterInfo
	mutex  sync.Mutex
	logger *log.Logger
}

type PulseMeterInfo struct {
	serno   uint32
	version uint32
	date    string
	time    string
}

func NewPulseTest(bus I2CBus, addr uint16, logger *log.Logger) *PulseTest {
	return &PulseTest{bus: bus, addr: addr, logger: logger}
}

// GetMeterInfo tries to populate PulseTest.meter from I2C.
// This is highly device-specific; this stub returns false unless you implement.
func (p *PulseTest) GetMeterInfo() bool {
	// TODO: Implement real I2C queries per QFAM protocol
	// Example:
	// w := []byte{ /* command */ }
	// r := make([]byte, 16)
	// if err := p.bus.Tx(p.addr, w, r); err != nil { return false }
	// parse r into p.meter...
	return false
}

// -------- Meter types, errors, helpers --------

var (
	ErrMeterContact = errors.New("could not contact meter")
	ErrBadReply     = errors.New("bad reply")
)

type Meter struct {
	serial           SerialPort
	i2c              I2CBus
	pulseTest        *PulseTest
	meterType        string // "TMX5", "TMX4", "QFAM", ...
	meterSerNo       uint32
	meterProtocol    string
	meterVersion     string
	meterDate        string
	meterTime        string
	numPhases        int
	useCurrentGating bool
	rBuf             []byte
	rBufMaxLen       int
	logger           *log.Logger
	// internal sync
	readMutex sync.Mutex
}

func NewMeter(serial SerialPort, i2c I2CBus, mType string, logger *log.Logger) (*Meter, error) {
	m := &Meter{
		serial:        serial,
		i2c:           i2c,
		meterType:     mType,
		meterProtocol: "unknown",
		rBufMaxLen:    20000,
		logger:        logger,
	}
	// For QFAM, set up PulseTest
	if mType == "QFAM" {
		m.pulseTest = NewPulseTest(i2c, 0x10, logger) // use real I2C addr
		if !m.pulseTest.GetMeterInfo() {
			return nil, ErrMeterContact
		}
		// populate fields from pulseTest.meter (example)
		if m.pulseTest.meter != nil {
			m.meterSerNo = m.pulseTest.meter.serno
			m.meterVersion = fmt.Sprintf("%08x", m.pulseTest.meter.version)
			// split date/time if available
		}
	} else {
		// For legacy meters, serial must be present
		if serial == nil {
			return nil, fmt.Errorf("serial required for meter type %s", mType)
		}
		// Basic login/setup
		if err := m.setupMeter(); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// -------- Serial command exchange --------

// sendCmd writes cmdStr to serial and reads until prompt regex or timeout.
// promptRE is a regular expression string to detect the end of reply.
func (m *Meter) sendCmd(cmdStr string, promptRE string, cmdTimeout time.Duration, comRetries int, noReply bool) (bool, error) {
	if comRetries <= 0 {
		comRetries = 1
	}
	if promptRE == "" {
		promptRE = `(CIP[\:>#\$\\]\r?$)|(S[\:>#\$\\]\r?$)`
	}
	promptRe := regexp.MustCompile(promptRE)

	for tries := 0; tries < comRetries; tries++ {
		if err := m.serial.FlushInput(); err != nil {
			m.logger.Printf("flush input error: %v", err)
		}
		m.rBuf = m.rBuf[:0]

		_, err := m.serial.Write([]byte(cmdStr))
		if err != nil {
			m.logger.Printf("serial write error: %v", err)
			continue
		}
		m.logger.Printf("sent: %s", cmdStr)

		if noReply {
			return true, nil
		}

		// read until timeout or match
		deadline := time.Now().Add(cmdTimeout)
		buf := make([]byte, 256)
		for {
			// adjust per-read timeout by setting serial read timeout via serial.SetReadTimeout before reading
			m.serial.SetReadTimeout(time.Millisecond * 200)
			n, err := m.serial.Read(buf)
			if err != nil {
				// non-fatal, retry loop
				// small sleep to avoid busy loop
				time.Sleep(50 * time.Millisecond)
			} else if n > 0 {
				m.rBuf = append(m.rBuf, buf[:n]...)
				if len(m.rBuf) > m.rBufMaxLen {
					return false, ErrBadReply
				}
				text := safeString(m.rBuf)
				m.logger.Printf("recv: %s", text)
				if promptRe.MatchString(text) {
					// success if no "error" in text
					if !regexp.MustCompile(`[Ee]rror`).MatchString(text) {
						return true, nil
					}
					return false, fmt.Errorf("remote error: %s", text)
				}
			}
			if time.Now().After(deadline) {
				break
			}
		}
		// retry
	}
	return false, ErrMeterContact
}

func safeString(b []byte) string {
	// replace invalid utf-8 with replacement char
	return string(b)
}

// -------- Setup and login --------

func (m *Meter) setupMeter() error {
	// Attempt login and version reading similar to Python setup_meter
	ok, err := m.login()
	if err != nil {
		return err
	}
	if !ok {
		return ErrMeterContact
	}

	// try to read version for some meter types
	if m.meterType == "TMX5" || m.meterType == "TMX5n" {
		// simple attempt to run "ver" and parse hex version
		if ok, _ := m.sendCmd("ver\r", `([0-9a-fA-F]{8})`, 2*time.Second, 2, false); ok {
			// extract first hex token
			re := regexp.MustCompile(`([0-9a-fA-F]{8})`)
			match := re.FindStringSubmatch(safeString(m.rBuf))
			if len(match) > 1 {
				m.meterVersion = match[1]
				// set protocol heuristics
				switch m.meterVersion[0] {
				case '2':
					m.meterProtocol = "TMX5B"
				case '3':
					m.meterProtocol = "TMX5C"
				case '5', '6':
					m.meterProtocol = "TMX5n"
				default:
					m.meterProtocol = "unknown"
				}
			}
		}
	}

	// set reasonable defaults if still unknown
	if m.meterProtocol == "" {
		m.meterProtocol = "unknown"
	}
	return nil
}

func (m *Meter) login() (bool, error) {
	// Simplified: try "attn -d\r" and accept success when prompt found.
	ok, err := m.sendCmd("attn -d\r", `(CIP[\:>#\$\\]\r?$)|(S[\:>#\$\\]\r?$)`, 2*time.Second, 2, false)
	if ok {
		return true, nil
	}
	// Fallbacks per meter type could be implemented here
	return false, err
}

// -------- KWH read logic (simplified) --------

func (m *Meter) GetKWH() ([]float64, error) {
	// Dispatch based on m.meterProtocol (simplified)
	switch m.meterProtocol {
	case "TMX5B", "TMX5C":
		return m.getKWH_TMX5B()
	case "TMX4", "TMX4M":
		return m.getKWH_TMX4()
	case "TMX5n":
		return m.getKWH_TMX5n()
	default:
		// Try TMX5B style as default
		return m.getKWH_TMX5B()
	}
}

func (m *Meter) getKWH_TMX5B() ([]float64, error) {
	if m.numPhases <= 0 {
		m.numPhases = 3 // default guess
	}
	if ok, err := m.sendCmd("md -Gt -Q2\r", `Mtr PH`, 3*time.Second, 3, false); !ok {
		return nil, err
	}
	reply := safeString(m.rBuf)
	lines := splitLines(reply)
	if len(lines) < 3 {
		return nil, ErrBadReply
	}
	var result []float64
	re := regexp.MustCompile(`( {1,8}\-?[\d\.]{1,17})`)
	for _, line := range lines[2 : 2+m.numPhases] {
		matches := re.FindAllString(line, -1)
		if len(matches) < 3 {
			return nil, ErrBadReply
		}
		// sum three columns (as in python code)
		var sum float64
		for i := 0; i < 3; i++ {
			v, err := strconv.ParseFloat(strings.TrimSpace(matches[i]), 64)
			if err != nil {
				return nil, err
			}
			sum += v
		}
		result = append(result, sum)
	}
	return result, nil
}

func (m *Meter) getKWH_TMX4() ([]float64, error) {
	// Similar to TMX5B but single-column extract
	if ok, err := m.sendCmd("md -Gt -Q2\r", `Mtr PH`, 3*time.Second, 3, false); !ok {
		return nil, err
	}
	reply := safeString(m.rBuf)
	lines := splitLines(reply)
	var result []float64
	re := regexp.MustCompile(`( {1,8}\-?[\d\.]{1,13})`)
	for _, line := range lines[2 : 2+m.numPhases] {
		matches := re.FindAllString(line, -1)
		if len(matches) < 1 {
			return nil, ErrBadReply
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(matches[0]), 64)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (m *Meter) getKWH_TMX5n() ([]float64, error) {
	// tmux5n multi-line formatting: try md -t
	if ok, err := m.sendCmd("md -t\r", `p# {7,12}kWH`, 3*time.Second, 3, false); !ok {
		return nil, err
	}
	reply := safeString(m.rBuf)
	lines := splitLines(reply)
	if len(lines) < 3 {
		return nil, ErrBadReply
	}
	var result []float64
	re := regexp.MustCompile(`(\-?[\d\.]{1,16})`)
	for i := 2; i < len(lines)-2; i += m.numPhases + 2 {
		line := lines[i]
		match := re.FindString(line)
		if match == "" {
			return nil, ErrBadReply
		}
		v, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

// -------- Phase diagnostics --------

func (m *Meter) GetPhaseData() ([][]float64, error) {
	switch m.meterProtocol {
	case "TMX4", "TMX4M":
		return m.getPhaseData_TMX4()
	case "TMX3M":
		return m.getPhaseData_TMX3M()
	default:
		return nil, fmt.Errorf("phase diagnostics not supported for protocol %s", m.meterProtocol)
	}
}

func (m *Meter) getPhaseData_TMX4() ([][]float64, error) {
	if ok, err := m.sendCmd("md -Gp\r", `Mtr PH`, 3*time.Second, 3, false); !ok {
		return nil, err
	}
	reply := safeString(m.rBuf)
	lines := splitLines(reply)
	if len(lines) < 3 {
		return nil, ErrBadReply
	}
	re := regexp.MustCompile(`( {0,8}\-?[\d\.]{1,13})([ muk])A( {0,8}\-?[\d\.]{1,13})([ muk])V( {0,8}\-?[\d\.]{1,13})([ muk])W( {0,8}\-?[\d\.]{1,13})([ muk])VAR`)
	var out [][]float64
	for _, line := range lines[2 : 2+m.numPhases] {
		match := re.FindStringSubmatch(line)
		if len(match) != 9 {
			return nil, ErrBadReply
		}
		vals := make([]float64, 4)
		for i := 0; i < 4; i++ {
			raw := strings.TrimSpace(match[1+2*i])
			unit := match[2+2*i]
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return nil, err
			}
			scale := unitScale(unit)
			vals[i] = v * scale
		}
		out = append(out, vals)
	}
	return out, nil
}

func (m *Meter) getPhaseData_TMX3M() ([][]float64, error) {
	// Emulate the Python behaviour: estimate amps by reading KWH twice, 30s apart.
	start, err := m.GetKWH()
	if err != nil {
		return nil, err
	}
	time.Sleep(30 * time.Second)
	end, err := m.GetKWH()
	if err != nil {
		return nil, err
	}
	if len(start) != len(end) {
		return nil, fmt.Errorf("kwh length mismatch")
	}
	out := make([][]float64, len(start))
	for i := range start {
		estimatedAmps := (end[i] - start[i]) * 200.0
		estimatedWatts := estimatedAmps * 120.0
		out[i] = []float64{estimatedAmps, 120.0, estimatedWatts, 0.0}
	}
	return out, nil
}

// -------- Utilities --------

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func unitScale(u string) float64 {
	switch u {
	case " ":
		return 1.0
	case "m":
		return 0.001
	case "u":
		return 0.000001
	case "k":
		return 1000.0
	default:
		return 1.0
	}
}

// -------- Example main skeleton (replace serial/i2c TODOs with periph implementations) --------

func main() {
	logger := log.Default()
	logger.Println("Meter Go skeleton starting")

	// TODO: Initialise periph host and create periph-based implementations of SerialPort and I2CBus.
	// Example (pseudo):
	// host.Init()
	// uart := periphOpenUART("/dev/ttyAMA0", 19200)
	// i2c := periphOpenI2C(1)
	// Use uart and i2c to create Meter.

	var serial SerialPort = nil // TODO: replace with real periph-based SerialPort
	var i2c I2CBus = nil        // TODO: replace with real periph-based I2CBus

	// For demonstration, fail fast if not implemented:
	if serial == nil {
		logger.Fatal("serial transport is not implemented. Implement SerialPort using periph.io and set here.")
	}

	meter, err := NewMeter(serial, i2c, "TMX5", logger)
	if err != nil {
		logger.Fatalf("NewMeter failed: %v", err)
	}

	kwh, err := meter.GetKWH()
	if err != nil {
		logger.Printf("GetKWH error: %v", err)
	} else {
		logger.Printf("KWH: %+v", kwh)
	}

	phase, err := meter.GetPhaseData()
	if err != nil {
		logger.Printf("GetPhaseData error: %v", err)
	} else {
		logger.Printf("Phase: %+v", phase)
	}
}
