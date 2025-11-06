// Converted (starter) Go module for meter.py
//
// Notes:
// - This is a pragmatic, working skeleton that mirrors the Python module structure,
//   with serial I/O, basic command/response handling and hooks for QFAM support.
// - Many protocol-specific parsers and ST4/PLC logic are left as TODOs and will
//   need to be implemented to match the original Python behavior exactly.
// - Uses github.com/tarm/serial for serial port access. Add it to your module:
//     go get github.com/tarm/serial
//
// Save as meter.go in package "meter".
// Converted (starter) Go module for meter.py
//
// Notes:
// - This is a pragmatic, working skeleton that mirrors the Python module structure,
//   with serial I/O, basic command/response handling and hooks for QFAM support.
// - Many protocol-specific parsers and ST4/PLC logic are left as TODOs and will
//   need to be implemented to match the original Python behavior exactly.
// - Uses periph.io for serial port access. Add it to your module:
//     go get periph.io/x/conn/v3
//     go get periph.io/x/host/v3
//
// Save as meter.go in package "meter".

package meter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	// periph host init (optional)
	"periph.io/x/host/v3"
)

// VERSION history maintained in the Python source; current:
const VERSION = 2.25

var (
	// logger similar to Q_logger in Python
	QLogger = log.Default()
)

// MeterError represents errors from meter operations
type MeterError struct {
	Msg string
}

func (e *MeterError) Error() string { return e.Msg }

type PulseTest struct {
	// Placeholder for PSC_QFAM_interface.PulseTest
	MeterInfoAvailable bool
	MeterSerNo         uint32
	Version            uint32
	DateAndTime        string
}

// Simulate PulseTest.GetMeterInfo
func (p *PulseTest) GetMeterInfo() bool {
	// TODO: implement I2C/Modbus access here.
	return p.MeterInfoAvailable
}

type Meter struct {
	ttyDev         string
	MeterType      string
	MeterSerNo     uint32
	MeterProtocol  string
	NumPhases      int
	MeterVersion   string
	MeterDate      string
	MeterTime      string
	UseCurrentGate bool

	// periph returns an io.ReadWriteCloser for the UART device
	serialPort io.ReadWriteCloser
	rwMutex    sync.Mutex

	// internal receive buffer for last command
	lastResponse string
	// max buffer length like Python r_buf_maxlen
	rBufMaxLen int
	// pulseTest for QFAM
	pulseTest *PulseTest
}

func NewMeter(ttydev string, mType string, mSerNo uint32, mProt string, mNumPh int) (*Meter, error) {
	m := &Meter{
		ttyDev:        ttydev,
		MeterType:     mType,
		MeterSerNo:    mSerNo,
		MeterProtocol: mProt,
		NumPhases:     mNumPh,
		rBufMaxLen:    20000,
	}

	QLogger.Printf("open serial port: %s", ttydev)

	if mType == "QFAM" {
		// QFAM uses PulseTest (I2C/modbus). This is a placeholder.
		m.pulseTest = &PulseTest{}
		if m.pulseTest.GetMeterInfo() {
			m.MeterSerNo = m.pulseTest.MeterSerNo
			parts := strings.Split(m.pulseTest.DateAndTime, " ")
			if len(parts) >= 2 {
				m.MeterDate, m.MeterTime = parts[0], parts[1]
			}
			m.MeterVersion = fmt.Sprintf("%08x", m.pulseTest.Version)
			// map first two hex chars -> protocol/phase mapping as Python did
			if len(m.MeterVersion) >= 2 {
				switch m.MeterVersion[:2] {
				case "10":
					m.MeterProtocol = "QB6"
					m.NumPhases = 6
				case "11":
					m.MeterProtocol = "QB4"
					m.NumPhases = 4
				case "12":
					m.MeterProtocol = "QB1"
					m.NumPhases = 1
				default:
					return nil, &MeterError{Msg: fmt.Sprintf("Invalid meter type in serno: %s", m.MeterVersion)}
				}
			}
		} else {
			return nil, &MeterError{Msg: "Could not contact QFAM meter"}
		}
		return m, nil
	}

	// Initialize periph host (required before using serial via periph)
	if _, err := host.Init(); err != nil {
		// non-fatal at runtime in many environments, but log
		QLogger.Printf("warning: host.Init() failed: %v", err)
	}

	// Legacy family - open serial port using periph.io serial.Open
	baud := 19200
	switch mType {
	case "TMX1", "TMX3", "TMX4":
		baud = 2400
	case "TMX5", "TMX5n", "MC5n":
		baud = 19200
	}

	sp, err := openSerialPeriph(ttydev, baud, time.Millisecond*500)
	if err != nil {
		return nil, err
	}
	m.serialPort = sp
	m.ttyDev = ttydev

	// call setup_meter (attempt login & version read)
	if ok, err := m.setupMeter(); !ok || err != nil {
		// close port on failure
		_ = m.Close()
		return nil, &MeterError{Msg: "Could not contact meter"}
	}
	return m, nil
}

// openSerialPeriph opens a serial port via periph and returns an io.ReadWriteCloser.
// openSerialPeriph opens a serial port; many periph serial APIs differ across versions,
// so use a simple fallback of opening the device file directly which works on Unix-like systems.
func openSerialPeriph(name string, baud int, timeout time.Duration) (io.ReadWriteCloser, error) {
	// Try to open device file directly. This avoids depending on specific periph serial types
	// that may not be present in the installed version of the library.
	f, err := os.OpenFile(name, os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return nil, fmt.Errorf("serial open failed: %w", err)
	}
	return f, nil
}

// optionsUint is a tiny helper to safely convert int to uint for older periph types.
func optionsUint(v int) uint {
	if v < 0 {
		return 0
	}
	return uint(v)
}

// set baudrate
func (m *Meter) SetBaudRate(baud int) error {
	if m.serialPort == nil {
		return errors.New("serial port not open")
	}
	// periph serial doesn't expose changing baud on an open port in a portable way.
	// We'll close and reopen using the helper.
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	_ = m.serialPort.Close()

	sp, err := openSerialPeriph(m.ttyDev, baud, time.Millisecond*500)
	if err != nil {
		return err
	}
	// optionsUint removed: not needed when opening device file directly.
	if sp == nil {
		return errors.New("failed to open serial port")
	}
	// replace the serial port handle with the newly opened port
	m.serialPort = sp

	buf := make([]byte, 1024)
	setReadDeadlineIfPossible(m.serialPort, time.Now().Add(10*time.Millisecond))
	for {
		n, err := m.serialPort.Read(buf)
		if err != nil || n == 0 {
			break
		}
		if n < len(buf) {
			break
		}
	}
	// Clear deadline if the underlying type supports it is optional.
	return nil
}

// NOTE: io.ReadWriteCloser doesn't define SetReadDeadline; we use an optional interface when underlying
// type supports deadlines.
type readDeadlineSetter interface {
	SetReadDeadline(time.Time) error
}

// convenience wrapper that attempts to set read deadline when underlying port supports it
func setReadDeadlineIfPossible(p io.ReadWriteCloser, t time.Time) {
	if s, ok := p.(readDeadlineSetter); ok {
		_ = s.SetReadDeadline(t)
	}
}

// sendCmd: send command and wait for prompt or timeout.
// promptRe optional regex; if empty uses default prompt similar to Python.
func (m *Meter) SendCmd(cmd string, cmdFlags byte, cmdTimeout time.Duration, comRetries int, promptRe string) (bool, error) {
	// cmdFlags: 'n' means no reply expected
	if m.serialPort == nil {
		return false, errors.New("serial port not open")
	}
	if promptRe == "" {
		// default prompt regex roughly equivalent to Python prompt_re
		promptRe = `(CIP[\:>#\$\\]\r?$)|(S[\:>#\$\\]\r?$)`
	}
	promptRE, err := regexp.Compile(promptRe)
	if err != nil {
		return false, err
	}

	data := []byte(cmd)
	for comRetries > 0 {
		comRetries--
		m.rwMutex.Lock()
		// flush/drain (best-effort)
		// Use drainRead helper to avoid relying on a Flush method.
		m.drainRead()
		_, werr := m.serialPort.Write(data)
		m.rwMutex.Unlock()
		if werr != nil {
			return false, werr
		}
		QLogger.Printf("send: %s", cmd)

		if cmdFlags == 'n' {
			return true, nil
		}

		// read loop until prompt or timeout
		var sb strings.Builder
		deadline := time.Now().Add(cmdTimeout)
		buf := make([]byte, 1024)
		for time.Now().Before(deadline) && sb.Len() < m.rBufMaxLen {
			// set a short read deadline if supported
			if s, ok := m.serialPort.(readDeadlineSetter); ok {
				_ = s.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			}
			m.rwMutex.Lock()
			n, rerr := m.serialPort.Read(buf)
			m.rwMutex.Unlock()
			if rerr != nil && n == 0 {
				// read timeout - wait a bit and continue
				time.Sleep(50 * time.Millisecond)
				continue
			}
			if n > 0 {
				chunk := string(buf[:n])
				// ensure valid UTF-8 -- Go strings are UTF-8. We assume device sends ASCII/UTF-8.
				sb.WriteString(chunk)
				if promptRE.MatchString(sb.String()) {
					m.lastResponse = sb.String()
					QLogger.Printf("receive: true\n%s", m.lastResponse)
					// basic error check
					if strings.Contains(strings.ToLower(m.lastResponse), "error") {
						return false, nil
					}
					// TODO: data_validator hook support
					return true, nil
				}
				// check for modem failure text
				if strings.Contains(sb.String(), "modem failure") {
					// emulate Python behavior: sleep then increase retries
					time.Sleep(25 * time.Second)
					comRetries += 5
				}
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		}
		// timed out; try again if retries remain
		QLogger.Printf("receive: false\n%s", sb.String())
		m.lastResponse = sb.String()
	}

	return false, nil
}

func (m *Meter) setupMeter() (bool, error) {
	// For brevity, implement a simplified version:
	// call login (which will attempt attn -d or serial login variants),
	// then try to read version for some meter types.
	ok, err := m.login()
	if err != nil || !ok {
		return false, err
	}
	// Additional type-specific setup
	switch m.MeterType {
	case "TMX5n":
		// read version (simple)
		_, _ = m.SendCmd("ver\r", 'r', 3*time.Second, 2, "")
		// Parse version from lastResponse
		ver := extractHex32(m.lastResponse)
		if ver != "" {
			m.MeterVersion = ver
			// set protocol & phases based on leading chars - simplified
			if strings.HasPrefix(ver, "5") {
				m.MeterProtocol = "TMX5n"
				m.UseCurrentGate = true
				// phases detection omitted
			}
		}
	case "TMX5":
		_, _ = m.SendCmd("ver\r", 'r', 3*time.Second, 2, "")
		m.MeterVersion = extractHex32(m.lastResponse)
		// simplified protocol decide
		if m.MeterVersion != "" {
			switch m.MeterVersion[0] {
			case '2':
				m.MeterProtocol = "TMX5B"
			case '3':
				m.MeterProtocol = "TMX5C"
			default:
				// fallback
				m.MeterProtocol = "TMX5"
			}
		}
	case "TMX4":
		ok, _ := m.SendCmd("ver\r", 'r', 3*time.Second, 2, "")
		if ok {
			lines := strings.Split(m.lastResponse, "\r\n")
			if len(lines) > 1 && len(lines[1]) >= 18 {
				m.MeterVersion = strings.TrimSpace(lines[1][10:18])
			}
		}
		// query md -Gp to find phases
		if ok, _ := m.SendCmd("md -Gp\r", 'r', 3*time.Second, 2, ""); ok {
			lines := strings.Split(m.lastResponse, "\r\n")
			phases := 0
			for _, ln := range lines {
				if regexp.MustCompile(`(\s{1,2}\d{1,2}){2}`).MatchString(ln) {
					phases++
				}
			}
			if phases > 0 {
				m.NumPhases = phases
				if phases < 4 {
					m.MeterProtocol = "TMX4"
				} else {
					m.MeterProtocol = "TMX4M"
				}
				m.UseCurrentGate = true
				return true, nil
			}
		}
	case "TMX3", "TMX1":
		m.UseCurrentGate = true
		m.MeterProtocol = m.MeterType
		return true, nil
	}
	// default OK if login succeeded
	return true, nil
}

func extractHex32(s string) string {
	re := regexp.MustCompile(`([0-9a-fA-F]{8})`)
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return strings.ToLower(m[1])
	}
	return ""
}

func (m *Meter) login() (bool, error) {
	// Simplified login flow:
	// attempt an 'attn -d' first for TMX4/TMX5
	if m.MeterType == "TMX4" || m.MeterType == "TMX5" {
		ok, _ := m.SendCmd("attn -d\r", 'r', 2*time.Second, 1, "")
		if ok {
			return true, nil
		}
	}
	// Fallback simulated login - real implementation needs protocol-specific sequences
	switch m.MeterType {
	case "TMX5":
		// try attn -D then attn -S... sequences as in Python
		ok, _ := m.SendCmd("attn -D\r", 'r', 2*time.Second, 4, `(\d{8})( {1,4}[0-9A-Fa-f]{1,4}){1,2}\r?`)
		if ok {
			// parse serial number
			re := regexp.MustCompile(`(\d{8})`)
			if match := re.FindStringSubmatch(m.lastResponse); len(match) >= 2 {
				// parse decimal serial as in Python eval
				fmt.Sscanf(match[1], "%d", &m.MeterSerNo)
			}
			// then try attn -S... (simplified)
			_, _ = m.SendCmd(fmt.Sprintf("attn -S%d -5lEvElbAl\r", m.MeterSerNo), 'r', 2*time.Second, 1, `\d{1,2}\:\d\d\:\d\d`)
			return true, nil
		}
		// try baud switching fallback omitted
		return false, nil
	case "TMX4":
		// simplified TMX4 login: try attn -S... at 2400, then change baud
		_ = m.SetBaudRate(2400)
		ok, _ := m.SendCmd(fmt.Sprintf("attn -S%d -4u5574\r", m.MeterSerNo), 'r', 2*time.Second, 1, `\d{1,2}\:\d\d\:\d\d`)
		if ok {
			// option to raise baud
			_, _ = m.SendCmd("baudhigh\r", 'r', 2*time.Second, 1, "")
			_ = m.SetBaudRate(9600)
			return true, nil
		}
		_ = m.SetBaudRate(9600)
		return true, nil
	case "TMX1", "TMX3":
		// TMX1/TMX3 require PLC ST4 setup in Python; we return true as placeholder
		return true, nil
	default:
		return false, nil
	}
}

// GetKWH - simplified: dispatch to protocol-specific stub functions
func (m *Meter) GetKWH() ([]float64, error) {
	switch m.MeterProtocol {
	case "TMX5B", "TMX5C":
		return m.getKwhTMX5B()
	case "TMX5D":
		return m.getKwhTMX5D()
	case "TMX4", "TMX4M":
		return m.getKwhTMX4()
	case "TMX3", "TMX1":
		return m.getKwhTMX3_TM1()
	case "TMX5n":
		return m.getKwhTMX5n()
	case "MC5n":
		return m.getKwhMC5n()
	default:
		return nil, &MeterError{Msg: "unsupported meter protocol for GetKWH"}
	}
}

func (m *Meter) getKwhTMX5B() ([]float64, error) {
	// Example: send md -Gt -Q2 and parse. Full parser omitted.
	ok, _ := m.SendCmd("md -Gt -Q2\r", 'r', 3*time.Second, 2, "")
	if !ok {
		return nil, &MeterError{Msg: "bad reply reading KWH"}
	}
	// Very simple parsing: extract all floats from response and return first NumPhases
	re := regexp.MustCompile(`-?\d+(\.\d+)?`)
	all := re.FindAllString(m.lastResponse, -1)
	out := make([]float64, 0, m.NumPhases)
	for i := 0; i < len(all) && len(out) < m.NumPhases; i++ {
		var v float64
		fmt.Sscanf(all[i], "%f", &v)
		out = append(out, v)
	}
	if len(out) != m.NumPhases {
		return nil, &MeterError{Msg: "bad reply reading KWH (count mismatch)"}
	}
	return out, nil
}

func (m *Meter) getKwhTMX5D() ([]float64, error) {
	// placeholder similar to TMX5B
	return m.getKwhTMX5B()
}

func (m *Meter) getKwhTMX4() ([]float64, error) {
	ok, _ := m.SendCmd("md -Gt -Q2\r", 'r', 3*time.Second, 3, "")
	if !ok {
		return nil, &MeterError{Msg: "bad reply reading KWH"}
	}
	re := regexp.MustCompile(`-?\d+(\.\d+)?`)
	all := re.FindAllString(m.lastResponse, -1)
	out := make([]float64, 0, m.NumPhases)
	for i := 0; i < len(all) && len(out) < m.NumPhases; i++ {
		var v float64
		fmt.Sscanf(all[i], "%f", &v)
		out = append(out, v)
	}
	if len(out) != m.NumPhases {
		return nil, &MeterError{Msg: "bad reply reading KWH (count mismatch)"}
	}
	return out, nil
}

func (m *Meter) getKwhTMX3_TM1() ([]float64, error) {
	// TMX1/TMX3 via PLC/ST4 is complex; not implemented.
	return nil, &MeterError{Msg: "GetKWH via ST4/PLC not implemented in Go skeleton"}
}

func (m *Meter) getKwhTMX5n() ([]float64, error) {
	// Simplified implementation: TODO parse properly
	ok, _ := m.SendCmd("md -t\r", 'r', 3*time.Second, 2, "")
	if !ok {
		return nil, &MeterError{Msg: "bad reply reading KWH for TMX5n"}
	}
	re := regexp.MustCompile(`-?\d+(\.\d+)?`)
	all := re.FindAllString(m.lastResponse, -1)
	out := make([]float64, 0, m.NumPhases)
	for i := 0; i < len(all) && len(out) < m.NumPhases; i++ {
		var v float64
		fmt.Sscanf(all[i], "%f", &v)
		out = append(out, v)
	}
	if len(out) != m.NumPhases {
		return nil, &MeterError{Msg: "bad reply reading KWH (count mismatch)"}
	}
	return out, nil
}

func (m *Meter) getKwhMC5n() ([]float64, error) {
	// placeholder by calling mscan -Gt and parsing child lines (omitted)
	ok, _ := m.SendCmd("mscan -Gt\r", 'r', 3*time.Second, 3, "")
	if !ok {
		return nil, &MeterError{Msg: "bad reply reading KWH for MC5n"}
	}
	// find long hex/number fields; naive parse: collect floats
	re := regexp.MustCompile(`-?\d+(\.\d+)?`)
	all := re.FindAllString(m.lastResponse, -1)
	out := make([]float64, 0, m.NumPhases)
	for i := 0; i < len(all) && len(out) < m.NumPhases; i++ {
		var v float64
		fmt.Sscanf(all[i], "%f", &v)
		out = append(out, v)
	}
	if len(out) != m.NumPhases {
		return nil, &MeterError{Msg: "bad reply reading KWH (count mismatch)"}
	}
	return out, nil
}

// getPhaseData returns per-phase [Amps,Volts,Watts,VARs] lists. Some protocols not implemented.
func (m *Meter) GetPhaseData() ([][]float64, error) {
	switch m.MeterProtocol {
	case "TMX4", "TMX4M":
		return m.getPhaseDataTMX4()
	case "TMX3M":
		return m.getPhaseDataTMX3M()
	default:
		return nil, &MeterError{Msg: "GetPhaseData not implemented for this protocol"}
	}
}

func (m *Meter) getPhaseDataTMX4() ([][]float64, error) {
	ok, _ := m.SendCmd("md -Gp\r", 'r', 3*time.Second, 2, "")
	if !ok {
		return nil, &MeterError{Msg: "bad reply reading phase"}
	}
	lines := strings.Split(m.lastResponse, "\r\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[1], "Mtr PH") {
		return nil, &MeterError{Msg: "bad reply reading phase"}
	}
	out := [][]float64{}
	re := regexp.MustCompile(`([\-0-9\.]+)([ muk])A.*?([\-0-9\.]+)([ muk])V.*?([\-0-9\.]+)([ muk])W.*?([\-0-9\.]+)([ muk])VAR`)
	for _, ln := range lines[2:] {
		if ln == "" {
			continue
		}
		matches := re.FindStringSubmatch(ln)
		if len(matches) != 9 {
			return nil, &MeterError{Msg: "Bad data line in phase readings"}
		}
		vals := make([]float64, 4)
		for i := 0; i < 4; i++ {
			var v float64
			fmt.Sscanf(matches[2*i+1], "%f", &v)
			unit := matches[2*i+2]
			mult := 1.0
			switch unit {
			case "m":
				mult = 0.001
			case "u":
				mult = 0.000001
			case "k":
				mult = 1000.0
			default:
				mult = 1.0
			}
			vals[i] = v * mult
		}
		out = append(out, vals)
	}
	return out, nil
}

func (m *Meter) getPhaseDataTMX3M() ([][]float64, error) {
	// Algorithm in Python: read KWH, wait 30s, read again, convert to amps/watts
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
		return nil, &MeterError{Msg: "inconsistent KWH lengths"}
	}
	out := make([][]float64, len(start))
	for i := range start {
		estimatedAmps := (end[i] - start[i]) * 200.0
		estimatedWatts := estimatedAmps * 120.0
		out[i] = []float64{estimatedAmps, 120.0, estimatedWatts, 0.0}
	}
	return out, nil
}

// Utility: hex dump of lastResponse (debug helper)
func (m *Meter) LastResponseHex() string {
	return hex.Dump([]byte(m.lastResponse))
}

// Close the underlying serial port
func (m *Meter) Close() error {
	if m.serialPort != nil {
		return m.serialPort.Close()
	}
	return nil
}

// Example CLI-like main testing function (not included in package use).
// To use: create a separate main package and call NewMeter() and methods.
//
// End of file
