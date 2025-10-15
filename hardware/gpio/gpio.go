// Converted from /home/jwashin/psctester/psctester/from_mike_20251014/cgi-bin/hardware_interface.py
// VERSION = 2.05      # Add support for QFAM
package gpio

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

const (
	VERSION = "2.05"
	SNAME   = "hardware_interface.go"

	// NOTE: Original Python used GPIO.BOARD numbering. go-rpio uses BCM numbering.
	// Keep the same numbers as in original file but you must adjust them to BCM
	// if you want to use real hardware.

	// Original physical pin numbers

	// AMPS_ON_RELAY      = 22
	// AMPS_HIGH_RELAY    = 11
	// AMPS_LOW_RELAY     = 16
	// VOLTS_ON_RELAY     = 15
	// RS485_2WIRE_RELAY  = 36
	// GATE_CONTROL       = 12
	// SUPPRESS_OPTO_GATE = 32
	// STATUS_LED         = 18
	// OVER_TEMP          = 31

	// converted to BCM (https://pinout.xyz/)

	AMPS_ON_RELAY      = 25
	AMPS_HIGH_RELAY    = 17
	AMPS_LOW_RELAY     = 23
	VOLTS_ON_RELAY     = 22
	RS485_2WIRE_RELAY  = 16
	GATE_CONTROL       = 18
	SUPPRESS_OPTO_GATE = 12
	STATUS_LED         = 24
	OVER_TEMP          = 6

	RELAY_DELAY = 100 * time.Millisecond
)

var QLogger = log.New(os.Stdout, "QLC_acc_tstr: ", log.LstdFlags)

func init() {
	f, err := os.OpenFile("./script_ver_work", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		fmt.Fprintf(f, "%s %s\n", SNAME, VERSION)
		f.Close()
	} else {
		QLogger.Printf("failed to write version file: %v", err)
	}
}

// GPIOAdapter abstracts GPIO operations so we can fallback to a mock when rpio is unavailable.
type GPIOAdapter interface {
	SetupOutput(pin int, initialHigh bool)
	SetupInputPullUp(pin int)
	Output(pin int, high bool)
	Input(pin int) bool
	Close()
}

type rpioAdapter struct {
	open bool
}

func newRPIOAdapter() (*rpioAdapter, error) {
	if err := rpio.Open(); err != nil {
		return nil, err
	}
	return &rpioAdapter{open: true}, nil
}

func (a *rpioAdapter) SetupOutput(pin int, initialHigh bool) {
	p := rpio.Pin(pin)
	p.Output()
	if initialHigh {
		p.High()
	} else {
		p.Low()
	}
}

func (a *rpioAdapter) SetupInputPullUp(pin int) {
	p := rpio.Pin(pin)
	p.Input()
	p.PullUp()
}

func (a *rpioAdapter) Output(pin int, high bool) {
	p := rpio.Pin(pin)
	if high {
		p.High()
	} else {
		p.Low()
	}
}

func (a *rpioAdapter) Input(pin int) bool {
	p := rpio.Pin(pin)
	return p.Read() == rpio.High
}

func (a *rpioAdapter) Close() {
	if a.open {
		rpio.Close()
		a.open = false
	}
}

// mockAdapter logs actions instead of touching hardware
type mockAdapter struct {
	state map[int]bool
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{state: make(map[int]bool)}
}

func (m *mockAdapter) SetupOutput(pin int, initialHigh bool) {
	QLogger.Printf("MOCK: SetupOutput pin %d initial %v", pin, initialHigh)
	m.state[pin] = initialHigh
}

func (m *mockAdapter) SetupInputPullUp(pin int) {
	QLogger.Printf("MOCK: SetupInputPullUp pin %d", pin)
	m.state[pin] = true // pulled up
}

func (m *mockAdapter) Output(pin int, high bool) {
	QLogger.Printf("MOCK: Output pin %d -> %v", pin, high)
	m.state[pin] = high
}

func (m *mockAdapter) Input(pin int) bool {
	v := m.state[pin]
	QLogger.Printf("MOCK: Input pin %d -> %v", pin, v)
	return v
}

func (m *mockAdapter) Close() {
	QLogger.Printf("MOCK: Close")
}

// TesterHardware mirrors Tester_hardware_obj in Python
type TesterHardware struct {
	gpio     GPIOAdapter
	testType string
}

func NewTesterHardware(testType string) *TesterHardware {
	var gpio GPIOAdapter
	r, err := newRPIOAdapter()
	if err != nil {
		QLogger.Printf("failed to open rpio (%v), using mock adapter", err)
		gpio = newMockAdapter()
	} else {
		gpio = r
	}

	// initialize pins as in original Python
	gpio.SetupOutput(AMPS_ON_RELAY, true)
	gpio.SetupOutput(AMPS_HIGH_RELAY, true)
	gpio.SetupOutput(AMPS_LOW_RELAY, true)
	gpio.SetupOutput(VOLTS_ON_RELAY, true)
	gpio.SetupOutput(GATE_CONTROL, true)
	gpio.SetupOutput(SUPPRESS_OPTO_GATE, true)
	gpio.SetupInputPullUp(OVER_TEMP)
	initialRS485 := false
	if testType == "QFAM" {
		initialRS485 = true
	}
	gpio.SetupOutput(RS485_2WIRE_RELAY, initialRS485)

	return &TesterHardware{
		gpio:     gpio,
		testType: testType,
	}
}

func (t *TesterHardware) SetAmpsOff() {
	QLogger.Println("Turn current off")
	t.gpio.Output(AMPS_ON_RELAY, true)
}

func (t *TesterHardware) SetAmpsOn() {
	QLogger.Println("Turn current on")
	t.gpio.Output(AMPS_ON_RELAY, false)
}

func (t *TesterHardware) PresetAmpsHigh() {
	t.SetAmpsOff()
	time.Sleep(RELAY_DELAY)
	QLogger.Println("Set up for high current")
	t.gpio.Output(AMPS_LOW_RELAY, false)
	t.gpio.Output(AMPS_HIGH_RELAY, false)
	time.Sleep(RELAY_DELAY)
}

func (t *TesterHardware) PresetAmpsLow() {
	t.SetAmpsOff()
	time.Sleep(RELAY_DELAY)
	QLogger.Println("Set up for low current")
	t.gpio.Output(AMPS_LOW_RELAY, false)
	t.gpio.Output(AMPS_HIGH_RELAY, true)
	time.Sleep(RELAY_DELAY)
}

func (t *TesterHardware) SetVoltsOff() {
	QLogger.Println("Turn voltage off")
	t.gpio.Output(VOLTS_ON_RELAY, true)
}

func (t *TesterHardware) SetVoltsOn() {
	QLogger.Println("Turn voltage on")
	t.gpio.Output(VOLTS_ON_RELAY, false)
}

func (t *TesterHardware) DoGatePulse(timeMs int) {
	QLogger.Println("Send gate pulse to Radian")
	t.gpio.Output(GATE_CONTROL, false)
	time.Sleep(time.Duration(timeMs) * time.Millisecond)
	t.gpio.Output(GATE_CONTROL, true)
}

func (t *TesterHardware) SuppressOptoGate() {
	QLogger.Println("Disable opto gate pulses")
	t.gpio.Output(SUPPRESS_OPTO_GATE, false)
}

func (t *TesterHardware) AllowOptoGate() {
	QLogger.Println("Enable opto gate pulses")
	t.gpio.Output(SUPPRESS_OPTO_GATE, true)
}

func (t *TesterHardware) DoCleanup() {
	t.gpio.Close()
}

func testing() {
	fmt.Println("start")

	TEST_TYPE := "TMX5"
	// If you want override test type, pass as first arg
	if len(os.Args) > 1 {
		TEST_TYPE = os.Args[1]
	}

	tester := NewTesterHardware(TEST_TYPE)

	keepRunning := true

	exitHwInterface := func() {
		keepRunning = false
	}

	setAmpsLow := func() {
		tester.PresetAmpsLow()
		tester.SetAmpsOn()
	}
	setAmpsHigh := func() {
		tester.PresetAmpsHigh()
		tester.SetAmpsOn()
	}

	options := map[string]func(){
		"amps off":                  tester.SetAmpsOff,
		"volts off":                 tester.SetVoltsOff,
		"volts on":                  tester.SetVoltsOn,
		"amps on low":               setAmpsLow,
		"amps on high":              setAmpsHigh,
		"Start Radian accumulation": func() { tester.DoGatePulse(50) },
		"exit":                      exitHwInterface,
	}

	fmt.Println("Choose one of the following options")
	for k := range options {
		fmt.Println(k)
	}

	reader := bufio.NewReader(os.Stdin)
	for keepRunning {
		fmt.Print("--> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			QLogger.Printf("read error: %v", err)
			break
		}
		choice := strings.TrimSpace(line)
		if fn, ok := options[choice]; ok {
			fn()
		} else {
			fmt.Println("not a valid option, please choose one of the following")
			for valid := range options {
				fmt.Println(valid)
			}
		}
	}

	tester.DoCleanup()
	fmt.Println("exit")
}
