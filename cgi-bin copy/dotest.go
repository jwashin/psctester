package main

// GitHub Copilot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const VERSION = "2.20"

var DOTEST_VERSION = VERSION

// Simple logger similar to Q_logger
var Qlogger *log.Logger

// Basic placeholder structs for hardware and meter interactions.
// These provide minimal behavior so the converted program compiles
// and produces a similar CSV/log structure. Replace with real
// implementations when integrating with actual hardware.

type Tester struct {
	Number int
}

func NewTester(testType string) *Tester {
	return &Tester{Number: 1}
}

func (t *Tester) SetVoltsOn()             {}
func (t *Tester) SetVoltsOff()            {}
func (t *Tester) PresetAmpsLow()          {}
func (t *Tester) PresetAmpsHigh()         {}
func (t *Tester) SetAmpsOn()              {}
func (t *Tester) SetAmpsOff()             {}
func (t *Tester) SuppressOptoGate()       {}
func (t *Tester) DoCleanup()              {}
func (t *Tester) SetFilename(name string) {}

type Radian struct {
	Model   string
	Serial  string
	Name    string
	Version string
	CalDate string
	Volts   float64
	amps    float64
}

func NewRadian(path string) *Radian {
	// placeholder values
	return &Radian{
		Model:   "RadianX",
		Serial:  "RAD12345",
		Name:    "Radian Standard",
		Version: "1.0",
		CalDate: "2023-01-01",
		Volts:   120.0,
		amps:    1.0,
	}
}

func (r *Radian) UpdateReadings() {
	// placeholder - nothing to do
}

func (r *Radian) GetKWH() (float64, error) {
	return 0.0, nil
}

type Meter struct {
	MeterSerNo       int
	NumPhases        int
	MeterProtocol    string
	ChannelMask      []bool
	UseCurrentGating bool
	MeterType        string
	NumActivePhases  int
	RawErrPercent    []float64
	ErrPercent       []float64
	PassFail         []string
	ChildData        [][]string
	PassedAllTests   bool
	MeterVersion     string
}

func NewMeter(port, testType string, serno int, protocol string, numPhases int) (*Meter, error) {
	m := &Meter{
		MeterSerNo:       serno,
		NumPhases:        max(1, numPhases),
		MeterProtocol:    protocol,
		ChannelMask:      make([]bool, max(1, numPhases)),
		UseCurrentGating: false,
		MeterType:        testType,
		MeterVersion:     "1.0.0",
	}
	for i := range m.ChannelMask {
		m.ChannelMask[i] = true
	}
	m.RawErrPercent = make([]float64, len(m.ChannelMask))
	m.ErrPercent = make([]float64, len(m.ChannelMask))
	m.PassFail = make([]string, len(m.ChannelMask))
	return m, nil
}

func (m *Meter) WaitSeconds(s int) {
	time.Sleep(time.Duration(s) * time.Second)
}

func (m *Meter) GetPhaseData() [][2]float64 {
	// return a slice of (amps, kwh) placeholder per phase
	out := make([][2]float64, m.NumPhases)
	for i := 0; i < m.NumPhases; i++ {
		out[i] = [2]float64{0.0, 0.0}
	}
	return out
}

func (m *Meter) GetKWH() ([]float64, error) {
	return []float64{0.0}, nil
}

func (m *Meter) SetupActiveChannelScan() bool {
	// placeholder returns true
	return true
}

// Test data structures
type PhaseDatum struct {
	TestSeconds float64
	RadianKWH   float64
	StartRead   float64
	EndRead     float64
	PercentErr  float64
}

type SingleTestLog struct {
	TestAmps   float64
	PhaseData  []PhaseDatum
	PercentUnc float64
}

type TestDataLog struct {
	Low  SingleTestLog
	High SingleTestLog
}

type Tests struct {
	Tester *Tester
	Radian *Radian
	DUT    *Meter
	Logger *log.Logger
}

func NewTests(t *Tester, r *Radian, d *Meter, l *log.Logger) *Tests {
	return &Tests{Tester: t, Radian: r, DUT: d, Logger: l}
}

func (t *Tests) DoKWHTest(which string, td *TestDataLog) {
	// Placeholder: simulate some results
	amps := 1.0
	if which == "high" {
		amps = 10.0
	}
	phases := max(1, t.DUT.NumPhases)
	sdat := SingleTestLog{
		TestAmps:   float64(amps),
		PhaseData:  make([]PhaseDatum, phases),
		PercentUnc: 0.5,
	}
	for i := 0; i < phases; i++ {
		sdat.PhaseData[i] = PhaseDatum{
			TestSeconds: 60.0,
			RadianKWH:   0.1,
			StartRead:   1000.0,
			EndRead:     1000.1,
			PercentErr:  0.0,
		}
		t.DUT.RawErrPercent[i] = 0.0
		t.DUT.ErrPercent[i] = 0.0
		t.DUT.PassFail[i] = "Pass"
	}
	if which == "low" {
		td.Low = sdat
	} else {
		td.High = sdat
	}
}

// Helpers
func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func writeScriptVerWork(name, version string) error {
	f, err := os.Create("./script_ver_work")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s %s\n", name, version)
	return err
}

func initLogger() {
	logPath := "/tmp/acc_test.log"
	// try to use /home/pi path; if not available, fallback to /tmp
	if _, err := os.Stat("/home/pi/acc_tester/log"); err == nil {
		logPath = "/home/pi/acc_tester/log/acc_test.log"
		os.MkdirAll("/home/pi/acc_tester/log", 0755)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// fallback to stdout logger
		Qlogger = log.New(os.Stdout, "", log.LstdFlags)
		Qlogger.Printf("Could not open log file %s: %v", logPath, err)
		return
	}
	Qlogger = log.New(f, "", log.LstdFlags)
}

func dotest() bool {
	filesloc := "./reports"
	ensureDir(filesloc)

	// write initial version log
	ioutil.WriteFile("do_log", []byte(fmt.Sprintf("starting version %s\n", VERSION)), 0644)

	// control.json check (if exists, validate)
	thefile := "../control.json"
	if _, err := os.Stat(thefile); err == nil {
		b, err := ioutil.ReadFile(thefile)
		if err != nil {
			return false
		}
		var tmp interface{}
		if err := json.Unmarshal(b, &tmp); err != nil {
			return false
		}
	}

	// append to do_log
	f, _ := os.OpenFile("do_log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintln(f, "file exists")
		f.Close()
	}

	// read JSON from stdin
	stdinBytes, _ := ioutil.ReadAll(os.Stdin)
	var indata map[string]interface{}
	if len(stdinBytes) > 0 {
		if err := json.Unmarshal(stdinBytes, &indata); err != nil {
			indata = nil
		}
	}
	if indata == nil {
		indata = map[string]interface{}{
			"serial":   "10130",
			"siteid":   "92240",
			"c":        "Qfam",
			"x3":       "MC3-24",
			"channels": []int{22, 23, 24},
		}
	}

	siteID := fmt.Sprintf("%v", indata["siteid"])
	serialStr := fmt.Sprintf("%v", indata["serial"])

	// infer test type
	testType := "UNKNOWN"
	meterSerNo := 0
	meterNumPhases := 0
	meterProtocol := "unknown"

	cval := strings.ToLower(fmt.Sprintf("%v", indata["c"]))
	switch cval {
	case "tmx1":
		testType = "TMX1"
		meterSerNo, _ = strconv.Atoi(serialStr)
		meterNumPhases = 3
		meterProtocol = "TMX1"
	case "tmx3":
		testType = "TMX3"
		meterSerNo, _ = strconv.Atoi(serialStr)
		x3 := fmt.Sprintf("%v", indata["x3"])
		if x3 == "RSM3" {
			meterNumPhases = 3
			meterProtocol = "TMX3"
		} else {
			meterNumPhases = 24
			meterProtocol = "TMX3M"
		}
	case "tmx4":
		testType = "TMX4"
		meterSerNo, _ = strconv.Atoi(serialStr)
	case "tmx5":
		testType = "TMX5"
	case "tmx5n":
		testType = "TMX5n"
	case "qfam", "QFAM":
		testType = "QFAM"
	default:
		testType = "QFAM"
	}

	Qlogger.Printf("Starting %s test", testType)

	tester := NewTester(testType)
	radian := NewRadian("/dev/tty_radian")

	if radian == nil {
		Qlogger.Println("Unable to contact Radian")
		time.Sleep(5 * time.Second)
		Qlogger.Println("done")
		return false
	}

	tester.SetVoltsOn()
	Qlogger.Println("Turning power on ...")
	time.Sleep(2 * time.Second)
	radian.UpdateReadings()
	Qlogger.Printf("Power on, %.2f volts", radian.Volts)
	if radian.Volts < 100.0 {
		Qlogger.Println("Unable to turn voltage on.  Make sure switch is in test position.")
		tester.SetFilename("no_file.csv")
		time.Sleep(1 * time.Second)
		Qlogger.Println("done")
		return true
	}

	// decide port list (placeholder)
	portList := []string{"/dev/tty_meter"}
	if testType == "TMX1" || testType == "TMX3" || testType == "TMX4" || testType == "QFAM" {
		portList = []string{"/dev/tty_rs485"}
	}

	var DUT *Meter
	okay := false
	var holdMessage string

	for _, tryPort := range portList {
		m, err := NewMeter(tryPort, testType, meterSerNo, meterProtocol, meterNumPhases)
		if err != nil {
			holdMessage = err.Error()
			continue
		}
		DUT = m
		okay = true
		break
	}

	if !okay {
		Qlogger.Println(holdMessage)
		return false
	}

	Qlogger.Printf("Found meter. Serial number: %d, %d Channels, %s", DUT.MeterSerNo, DUT.NumPhases, DUT.MeterProtocol)
	// meterUpdateDelay := 25

	// channel mask defaults already set in NewMeter
	if DUT.MeterProtocol == "TMX4M" {
		Qlogger.Println("Found MC-4.  (active channel check placeholder)")
		// placeholder for active channel check
	}

	if DUT.MeterProtocol == "TMX3M" {
		for i := range DUT.ChannelMask {
			DUT.ChannelMask[i] = false
		}
		if arr, ok := indata["channels"].([]interface{}); ok {
			for _, chv := range arr {
				ch := int(chv.(float64))
				if ch-1 >= 0 && ch-1 < len(DUT.ChannelMask) {
					DUT.ChannelMask[ch-1] = true
				}
			}
		}
		okay = DUT.SetupActiveChannelScan()
	}

	if !okay {
		Qlogger.Println("No active channels detected or setup failed")
		return false
	}

	tester.PresetAmpsLow()
	if DUT.UseCurrentGating {
		tester.SuppressOptoGate()
	} else {
		tester.SetAmpsOn()
	}

	runDatetime := time.Now()
	yearValid := runDatetime.Format("2006")
	var numPoints int
	if DUT.MeterType == "QFAM" {
		numPoints = DUT.NumPhases
	} else if DUT.NumPhases <= 3 {
		numPoints = 1
	} else {
		numPoints = DUT.NumPhases / 2
	}

	filename := fmt.Sprintf("%s_%s_%d.csv", siteID, runDatetime.Format("20060102_150405"), DUT.MeterSerNo)
	Qlogger.Printf("Log file name: %s", filename)
	tester.SetFilename(filename)

	logfilePath := filepath.Join(filesloc, filename)
	lf, err := os.Create(logfilePath)
	if err != nil {
		Qlogger.Printf("Failed to create log file %s: %v", logfilePath, err)
		return false
	}
	defer lf.Close()

	// write header similar to python output
	fmt.Fprintf(lf, "Date and time of accuracy tests for,,,,, %s,,%s,test point%s\n", siteID, yearValid, func() string {
		if numPoints == 1 {
			return ""
		}
		return "s"
	}())
	fmt.Fprintf(lf, "%s,,%s\n", runDatetime.Format("01/02/2006"), runDatetime.Format("15:04:05"))
	fmt.Fprintf(lf, "\n\nKWH standard used:,,,,,tester #%d\n", tester.Number)
	fmt.Fprintf(lf, "Radian Model:,,%s\n", radian.Model)
	fmt.Fprintf(lf, "Serial Number:,,%s\n", radian.Serial)
	fmt.Fprintf(lf, "Version Number:,,%s\n", radian.Version)
	fmt.Fprintf(lf, "Last Calibrated:,,%s\n\n", radian.CalDate)

	fmt.Fprintf(lf, "Meter under test:\n")
	fmt.Fprintf(lf, "Serial Number:,,%d\n", DUT.MeterSerNo)
	fmt.Fprintf(lf, "Software Version:,,%s\n", DUT.MeterVersion)
	fmt.Fprintf(lf, "# of channels:,,%d\n\n\n\n", DUT.NumPhases)

	tests := NewTests(tester, radian, DUT, Qlogger)
	DUT.PassedAllTests = true
	active := 0
	for _, v := range DUT.ChannelMask {
		if v {
			active++
		}
	}
	DUT.NumActivePhases = active

	td := &TestDataLog{}

	// run tests with placeholders
	defer func() {
		// ensure amps off and cleanup
		tester.SetAmpsOff()
		tester.SetVoltsOff()
		tester.DoCleanup()
		// write script_ver_work file
		writeScriptVerWork("dotest.py", VERSION)
		cmd := exec.Command("cp", "-p", "./script_ver_work", "../script_ver")
		_ = cmd.Run()
		time.Sleep(2 * time.Second)
		Qlogger.Println("done")
	}()

	Qlogger.Println("Setting up low-current test.")
	tests.DoKWHTest("low", td)

	fmt.Fprintln(lf, "Accuracy test results:")
	// log KWH results (low)
	fmt.Fprintln(lf, "\nLow Current Test:")
	if len(DUT.ChildData) == 0 {
		fmt.Fprintln(lf, "Channel,Percent Error,,Within tolerance?")
		for idx := range DUT.ErrPercent {
			if DUT.ChannelMask[idx] {
				fmt.Fprintf(lf, "%d,%.2f,,%s\n", idx+1, DUT.ErrPercent[idx], DUT.PassFail[idx])
			}
		}
	} else {
		fmt.Fprintln(lf, "Channel,Percent Error,,Within tolerance?,,,Child Serno,,Child channel")
		for idx := range DUT.ErrPercent {
			if DUT.ChannelMask[idx] {
				child := ""
				if idx < len(DUT.ChildData) && len(DUT.ChildData[idx]) >= 3 {
					child = fmt.Sprintf("%s,%s", DUT.ChildData[idx][1], DUT.ChildData[idx][2])
				}
				fmt.Fprintf(lf, "%d,%.2f,,%s,,,%s\n", idx+1, DUT.ErrPercent[idx], DUT.PassFail[idx], child)
			}
		}
	}
	// high test
	Qlogger.Println("Setting up high-current test.")
	tests.DoKWHTest("high", td)
	tester.SetAmpsOff()

	fmt.Fprintln(lf, "\nHigh Current Test:")
	if len(DUT.ChildData) == 0 {
		fmt.Fprintln(lf, "Channel,Percent Error,,Within tolerance?")
		for idx := range DUT.ErrPercent {
			if DUT.ChannelMask[idx] {
				fmt.Fprintf(lf, "%d,%.2f,,%s\n", idx+1, DUT.ErrPercent[idx], DUT.PassFail[idx])
			}
		}
	} else {
		fmt.Fprintln(lf, "Channel,Percent Error,,Within tolerance?,,,Child Serno,,Child channel")
		for idx := range DUT.ErrPercent {
			if DUT.ChannelMask[idx] {
				child := ""
				if idx < len(DUT.ChildData) && len(DUT.ChildData[idx]) >= 3 {
					child = fmt.Sprintf("%s,%s", DUT.ChildData[idx][1], DUT.ChildData[idx][2])
				}
				fmt.Fprintf(lf, "%d,%.2f,,%s,,,%s\n", idx+1, DUT.ErrPercent[idx], DUT.PassFail[idx], child)
			}
		}
	}

	workStatus := "Pass"
	if !DUT.PassedAllTests {
		workStatus = "Fail"
	}

	Qlogger.Printf("test done.  Meter %sed.", workStatus)
	fmt.Fprintf(lf, "\n\nTest conclusion:,,%s\n\n\n\n", workStatus)

	// log test details
	fmt.Fprintln(lf, "Test data details:")
	fmt.Fprintln(lf, "Amps,,Seconds,,actual KWH,,Start read,,End read,,Error %,,Uncertainty")
	// low
	for _, pd := range td.Low.PhaseData {
		fmt.Fprintf(lf, "%v,,%v,,%v,,%v,,%v,,%v\n", td.Low.TestAmps, pd.TestSeconds, pd.RadianKWH, pd.StartRead, pd.EndRead, td.Low.PercentUnc)
	}
	// high
	for _, pd := range td.High.PhaseData {
		fmt.Fprintf(lf, "%v,,%v,,%v,,%v,,%v,,%v\n", td.High.TestAmps, pd.TestSeconds, pd.RadianKWH, pd.StartRead, pd.EndRead, td.High.PercentUnc)
	}

	lf.Sync()

	return true
}

func main() {
	initLogger()
	Qlogger.Printf("Starting do_test() Version %s on tester# %d", DOTEST_VERSION, 1)

	s := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				Qlogger.Printf("Script failed: %v", r)
				s = false
			}
		}()
		s = dotest()
	}()

	// CGI-ish output similar to original script
	fmt.Println("Content-type: text/plain\n")
	fmt.Printf("%v\n", s)
}
