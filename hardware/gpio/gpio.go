// Converted from /home/jwashin/psctester/psctester/from_mike_20251014/cgi-bin/hardware_interface.py
// VERSION = 2.05      # Add support for QFAM
// Rewritten to use periph.io instead of go-rpio
package gpio

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	VERSION = "2.05"
	SNAME   = "hardware_interface.py"

	// NOTE: Original Python used GPIO.BOARD numbering. periph/gpioreg uses
	// Linux / BCM style names like "GPIO25" when running on a Raspberry Pi.
	// Keep the same names as in original file but you must adjust them to the
	// names your system exposes if different.

	AMPS_ON_RELAY      = "GPIO25"
	AMPS_HIGH_RELAY    = "GPIO17"
	AMPS_LOW_RELAY     = "GPIO23"
	VOLTS_ON_RELAY     = "GPIO22"
	RS485_2WIRE_RELAY  = "GPIO16"
	GATE_CONTROL       = "GPIO18"
	SUPPRESS_OPTO_GATE = "GPIO12"
	STATUS_LED         = "GPIO24"
	OVER_TEMP          = "GPIO6"

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

// GPIOAdapter abstracts GPIO operations so we can fallback to a mock when periph is unavailable.
type GPIOAdapter interface {
	SetupOutput(pin string, initialHigh bool)
	SetupInputPullUp(pin string)
	Output(pin string, high bool)
	Input(pin string) bool
	Close()
}

// periphAdapter implements GPIOAdapter using periph.io
type periphAdapter struct {
	// pins caches opened pins by name
	pins map[string]gpio.PinIO
	// initialized indicates host.Init() succeeded
	initialized bool
}

func newPeriphAdapter() (*periphAdapter, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}
	return &periphAdapter{
		pins:        make(map[string]gpio.PinIO),
		initialized: true,
	}, nil
}

func (a *periphAdapter) getPin(name string) gpio.PinIO {
	if p, ok := a.pins[name]; ok {
		return p
	}
	p := gpioreg.ByName(name)
	if p == nil {
		QLogger.Printf("periph: pin not found: %s", name)
		return nil
	}
	a.pins[name] = p
	return p
}

func (a *periphAdapter) SetupOutput(pin string, initialHigh bool) {
	p := a.getPin(pin)
	if p == nil {
		return
	}
	if initialHigh {
		if err := p.Out(gpio.High); err != nil {
			QLogger.Printf("periph: failed to set %s high: %v", pin, err)
		}
	} else {
		if err := p.Out(gpio.Low); err != nil {
			QLogger.Printf("periph: failed to set %s low: %v", pin, err)
		}
	}
}

func (a *periphAdapter) SetupInputPullUp(pin string) {
	p := a.getPin(pin)
	if p == nil {
		return
	}
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		QLogger.Printf("periph: failed to configure %s as input pull-up: %v", pin, err)
	}
}

func (a *periphAdapter) Output(pin string, high bool) {
	p := a.getPin(pin)
	if p == nil {
		return
	}
	if high {
		if err := p.Out(gpio.High); err != nil {
			QLogger.Printf("periph: Output(%s, high) error: %v", pin, err)
		}
	} else {
		if err := p.Out(gpio.Low); err != nil {
			QLogger.Printf("periph: Output(%s, low) error: %v", pin, err)
		}
	}
}

func (a *periphAdapter) Input(pin string) bool {
	p := a.getPin(pin)
	if p == nil {
		return false
	}
	return p.Read() == gpio.High
}

func (a *periphAdapter) Close() {
	// periph pins don't need explicit close. Clear cache.
	a.pins = nil
	a.initialized = false
}

// mockAdapter logs actions instead of touching hardware
type mockAdapter struct {
	state map[string]bool
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{state: make(map[string]bool)}
}

func (m *mockAdapter) SetupOutput(pin string, initialHigh bool) {
	QLogger.Printf("MOCK: SetupOutput pin %s initial %v", pin, initialHigh)
	m.state[pin] = initialHigh
}

func (m *mockAdapter) SetupInputPullUp(pin string) {
	QLogger.Printf("MOCK: SetupInputPullUp pin %s", pin)
	m.state[pin] = true // pulled up
}

func (m *mockAdapter) Output(pin string, high bool) {
	QLogger.Printf("MOCK: Output pin %s -> %v", pin, high)
	m.state[pin] = high
}

func (m *mockAdapter) Input(pin string) bool {
	v := m.state[pin]
	QLogger.Printf("MOCK: Input pin %s -> %v", pin, v)
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
	var gpioDev GPIOAdapter
	per, err := newPeriphAdapter()
	if err != nil {
		QLogger.Printf("failed to initialize periph (%v), using mock adapter", err)
		gpioDev = newMockAdapter()
	} else {
		gpioDev = per
	}

	// initialize pins as in original Python
	gpioDev.SetupOutput(AMPS_ON_RELAY, true)
	gpioDev.SetupOutput(AMPS_HIGH_RELAY, true)
	gpioDev.SetupOutput(AMPS_LOW_RELAY, true)
	gpioDev.SetupOutput(VOLTS_ON_RELAY, true)
	gpioDev.SetupOutput(GATE_CONTROL, true)
	gpioDev.SetupOutput(SUPPRESS_OPTO_GATE, true)
	gpioDev.SetupInputPullUp(OVER_TEMP)
	initialRS485 := false
	if testType == "QFAM" {
		initialRS485 = true
	}
	gpioDev.SetupOutput(RS485_2WIRE_RELAY, initialRS485)

	return &TesterHardware{
		gpio:     gpioDev,
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
