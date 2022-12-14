package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var filesloc = ""
var archives = ""

func makeReportDirs() {
	curdir, _ := os.Getwd()
	sp := strings.Split(curdir, "/")
	nl := []string{}
	found := false
	for k, v := range sp {
		if v == "home" {
			found = true
			nl = append(nl, v)
			nl = append(nl, sp[k+1])
			break
		}
		if !found {
			nl = append(nl, v)
		}
	}
	homedir := "/" + filepath.Join(nl...)
	filesloc = filepath.Join(homedir, "acc_tester", "reports")
	archives = filepath.Join(homedir, "acc_tester", "archive")
	os.MkdirAll(filesloc, 0755)
	os.MkdirAll(archives, 0755)
}

func main() {
	assureControlFile()
	makeReportDirs()

	r := gin.Default()

	wwwroot := "./build"

	_, err := os.Stat("build")
	if errors.Is(err, os.ErrNotExist) {
		wwwroot = ""
	}

	// index
	r.StaticFile("/", filepath.Join(wwwroot, "index.html"))
	r.StaticFile("/index.html", filepath.Join(wwwroot, "index.html"))

	// served folders
	r.Static("/media", filepath.Join(wwwroot, "media"))
	r.Static("/packages", filepath.Join(wwwroot, "packages"))

	// web app
	r.StaticFile("/psctester.css", filepath.Join(wwwroot, "psctester.css"))
	r.StaticFile("/psctester.dart.js", filepath.Join(wwwroot, "psctester.dart.js"))
	r.StaticFile("script_ver", "./script_ver")
	r.StaticFile("version_tag.txt", "./version_tag.txt")

	// web app development files
	r.StaticFile("/psctester.dart", filepath.Join(wwwroot, "psctester.dart"))
	r.StaticFile("/psctester.dart.bootstrap.js", filepath.Join(wwwroot, "psctester.dart.bootstrap.js"))
	r.StaticFile("/psctester.sound.ddc.js", filepath.Join(wwwroot, "psctester.sound.ddc.js"))
	r.StaticFile("/psctester.sound.ddc.js.map", filepath.Join(wwwroot, "psctester.sound.ddc.js.map"))

	r.GET("/control.json",
		func(c *gin.Context) {
			d, _ := os.ReadFile("control.json")
			c.String(200, "%s", string(d))
		})

	r.POST("cgi-bin/getfile.py",
		func(c *gin.Context) {
			filename := c.Query("filename")
			t, err := os.ReadFile(filepath.Join(filesloc, filename))
			if err == nil {
				c.Header("Content-Disposition", "attachment; filename="+filename)
				c.String(http.StatusOK, "%s", string(t))
			} else {
				c.String(http.StatusNotFound, err.Error())
			}
		})

	r.POST("/cgi-bin/archive_files.py",
		func(c *gin.Context) {
			r := c.PostFormArray("files")
			outs := []string{}
			for _, file := range r {
				oldfile := filepath.Join(filesloc, file)
				newfile := filepath.Join(archives, file)
				err := os.Rename(oldfile, newfile)
				if err != nil {
					outs = append(outs, err.Error())
				}
			}
			if len(outs) > 0 {
				c.String(http.StatusInternalServerError, "%s", strings.Join(outs, ";"))
			}
			c.String(http.StatusOK, "Archived: %s", strings.Join(r, ";"))
		})

	r.POST("/cgi-bin/get_zip.py",
		func(c *gin.Context) {
			r := c.PostFormArray("files")
			outs := []string{}

			buf := new(bytes.Buffer)
			w := zip.NewWriter(buf)

			for _, v := range r {
				file := filepath.Join(filesloc, v)
				data, err := os.ReadFile(file)
				if err != nil {
					outs = append(outs, err.Error())
					break
				}
				f, err2 := w.Create(v)
				if err2 != nil {
					outs = append(outs, err2.Error())
					break
				}
				f.Write(data)
			}
			w.Close()

			if len(outs) > 0 {
				c.String(http.StatusInternalServerError, "%s", strings.Join(outs, ";"))
			}
			date := time.Now().Format(time.RFC3339)
			c.Header("Content-Disposition", "attachment; filename=acc_"+date+".zip")
			c.Data(http.StatusOK, "application/zip", buf.Bytes())
		})

	r.GET("/cgi-bin/ip_addr.py",
		func(c *gin.Context) {
			cmd := exec.Command("python3", "./cgi-bin/ip_addr.py")
			cmdReader, _ := cmd.StdoutPipe()
			scanner := bufio.NewScanner(cmdReader)
			done := make(chan string)
			go func() {
				tst := ""
				for scanner.Scan() {
					tst += scanner.Text() + "\n"
				}
				done <- tst
			}()
			cmd.Start()
			s := <-done
			err := cmd.Wait()
			if err != nil {
				c.String(http.StatusInternalServerError, "could not obtain ip address")
			}
			s = strings.ReplaceAll(s, "Content-type: text/json\n\n", "")
			c.String(http.StatusOK, "%s", s)
		})

	r.GET("/cgi-bin/fileslist.py",
		func(c *gin.Context) {
			cmd := exec.Command("python3", "./cgi-bin/fileslist.py")
			cmdReader, _ := cmd.StdoutPipe()
			scanner := bufio.NewScanner(cmdReader)
			done := make(chan string)
			go func() {
				tst := ""
				for scanner.Scan() {
					tst += scanner.Text() + "\n"
				}
				done <- tst
			}()
			cmd.Start()
			s := <-done
			err := cmd.Wait()
			if err != nil {
				c.String(http.StatusInternalServerError, "file acquisition error")
			}
			s = strings.ReplaceAll(s, "Content-type: text/json\n\n", "")
			c.String(http.StatusOK, "%s", s)
		})

	r.POST("cgi-bin/dotest.py",
		func(c *gin.Context) {
			filename := "./cgi-bin/dotest.py"
			if os.Getenv("PSC_TRIAL") != "" {
				filename = "./cgi-bin/jim_dotest.py"
			}
			cmd := exec.Command("python3", filename)
			buffer := bytes.Buffer{}
			d := Indata{}
			c.BindJSON(&d)
			st, _ := json.Marshal(&d)
			buffer.Write(st)
			cmd.Stdin = &buffer
			cmdReader, _ := cmd.StdoutPipe()
			scanner := bufio.NewScanner(cmdReader)
			done := make(chan string)
			go func() {
				tst := ""
				for scanner.Scan() {
					tst += scanner.Text() + "\n"
				}
				done <- tst
			}()
			cmd.Start()
			s := <-done
			err := cmd.Wait()
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			}
			s = strings.ReplaceAll(s, "Content-type: text/plain\n", "")
			s = strings.ReplaceAll(s, "Content-type: application/json\n", "")
			c.String(http.StatusOK, "%s", s)
		})

	r.GET("/cgi-bin/initcontrolfile.py",
		func(c *gin.Context) {
			cmd := exec.Command("python3", "./cgi-bin/initcontrolfile.py")
			cmdReader, _ := cmd.StdoutPipe()
			scanner := bufio.NewScanner(cmdReader)
			done := make(chan string)
			go func() {
				tst := ""
				for scanner.Scan() {
					tst += scanner.Text() + "\n"
				}
				done <- tst
			}()
			cmd.Start()
			s := <-done
			err := cmd.Wait()
			if err != nil {
				c.String(http.StatusInternalServerError, "could not initialize file")
			}
			c.String(http.StatusOK, "%s", s)
		})

	r.GET("cgi-bin/getdate.py",
		func(c *gin.Context) {
			s := time.Now().Format("2006-01-02T15:04:05-0700")
			// s = s + "Z"
			c.JSON(http.StatusOK, gin.H{"datetime": s})
		})

	r.POST("cgi-bin/settime.py",
		func(c *gin.Context) {
			var data DateList
			c.BindJSON(&data)
			SetSystemDate(makeDate(data))
			c.JSON(http.StatusOK, gin.H{"resp": true})
		})

	r.GET("/flutter_service_worker.js", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	port := "8080"
	if os.Getenv("PSC_PORT") != "" {
		port = os.Getenv("PSC_PORT")
	}

	r.Run(":" + port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type ControlFile struct {
	Filename string   `json:"filename"`
	Version  string   `json:"version"`
	Start    string   `json:"start"`
	Status   string   `json:"status"`
	Serial   string   `json:"serial"`
	End      string   `json:"end"`
	Messages []string `json:"messages"`
}

type Indata struct {
	Serial   string `json:"serial"`
	Siteid   string `json:"siteid"`
	C        string `json:"c"`
	X3       string `json:"x3"`
	Channels []int  `json:"channels"`
}

type DateList struct {
	DT []int `json:"dt"`
}

func SetSystemDate(newTime time.Time) error {
	_, lookErr := exec.LookPath("date")
	if lookErr != nil {
		fmt.Printf("Date binary not found, cannot set system date: %s\n", lookErr.Error())
		return lookErr
	} else {
		//dateString := newTime.Format("2006-01-2 15:4:5")
		dateString := newTime.Format("2 Jan 2006 15:04:05")
		fmt.Printf("Setting system date to: %s\n", dateString)
		args := []string{"--set", dateString}
		return exec.Command("date", args...).Run()
	}
}
func makeDate(dt DateList) time.Time {
	// List s = [t.year, t.month, t.day, t.hour, t.minute, t.second, t.millisecond];
	sp := dt.DT
	y := sp[0]
	m := sp[1]
	d := sp[2]
	hr := sp[3]
	min := sp[4]
	sec := sp[5]
	msec := sp[6]
	month := time.January
	switch m {
	case 1:
		month = time.January
	case 2:
		month = time.February
	case 3:
		month = time.March
	case 4:
		month = time.April
	case 5:
		month = time.May
	case 6:
		month = time.June
	case 7:
		month = time.July
	case 8:
		month = time.August
	case 9:
		month = time.September
	case 10:
		month = time.October
	case 11:
		month = time.November
	case 12:
		month = time.December
	}
	return time.Date(y, month, d, hr, min, sec, msec, time.UTC)
}

func assureControlFile() {
	os.WriteFile("control.json", []byte("{}"), 0644)
}
