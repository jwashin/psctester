package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

	mux := http.NewServeMux()

	mime.AddExtensionType(".dart.js", "text/javascript")
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".css", "text/css")

	wwwroot := "./build"

	_, err := os.Stat("build")
	if errors.Is(err, os.ErrNotExist) {
		wwwroot = ""
	}

	// index
	mux.Handle("/", http.FileServer(http.Dir(wwwroot)))

	// // static files in www root
	// for _, v := range []string{
	// 	"index.html",
	// 	"favicon.ico",
	// 	"favicon-32x32.png",
	// 	"favicon-16x16.png",
	// 	"apple-touch-icon.png",
	// 	"android-chrome-512x512.png",
	// 	"android-chrome-192x192.png",
	// 	"site.webmanifest",
	// 	"about.txt",
	// } {
	// 	route := "/" + v
	// 	file := filepath.Join(wwwroot, v)
	// 	mux.HandleFunc(route, func(file string) http.HandlerFunc {
	// 		return func(w http.ResponseWriter, r *http.Request) {
	// 			http.ServeFile(w, r, file)
	// 		}
	// 	}(file))
	// }

	// served folders
	mediaDir := filepath.Join(wwwroot, "media")
	packagesDir := filepath.Join(wwwroot, "packages")
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(mediaDir))))
	mux.Handle("/packages/", http.StripPrefix("/packages/", http.FileServer(http.Dir(packagesDir))))

	// web app
	// mux.HandleFunc("/psctester.css", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.css"))
	// })
	// mux.HandleFunc("/psctester.dart.js", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.dart.js"))
	// })
	mux.HandleFunc("/script_ver", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./script_ver")
	})
	mux.HandleFunc("/version_tag.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./version_tag.txt")
	})

	// web app development files
	// mux.HandleFunc("/psctester.dart", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.dart"))
	// })
	// mux.HandleFunc("/psctester.dart.js.deps", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.dart.js.deps"))
	// })
	// mux.HandleFunc("/psctester.sound.ddc.js", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.sound.ddc.js"))
	// })
	// mux.HandleFunc("/psctester.dart.js.map", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, filepath.Join(wwwroot, "psctester.dart.js.map"))
	// })

	mux.HandleFunc("/control.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		d, _ := os.ReadFile("control.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(d)
	})

	mux.HandleFunc("/cgi-bin/getfile.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := r.URL.Query().Get("filename")
		t, err := os.ReadFile(filepath.Join(filesloc, filename))
		if err == nil {
			w.Header().Set("Content-Disposition", "attachment; filename="+filename)
			w.WriteHeader(http.StatusOK)
			w.Write(t)
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	})

	mux.HandleFunc("/cgi-bin/archive_files.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		files := r.Form["files"]
		outs := []string{}
		for _, file := range files {
			oldfile := filepath.Join(filesloc, file)
			newfile := filepath.Join(archives, file)
			err := os.Rename(oldfile, newfile)
			if err != nil {
				outs = append(outs, err.Error())
			}
		}
		if len(outs) > 0 {
			http.Error(w, strings.Join(outs, ";"), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Archived: " + strings.Join(files, ";")))
	})

	mux.HandleFunc("/cgi-bin/get_zip.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		files := r.Form["files"]
		outs := []string{}

		buf := new(bytes.Buffer)
		wr := zip.NewWriter(buf)

		for _, v := range files {
			file := filepath.Join(filesloc, v)
			data, err := os.ReadFile(file)
			if err != nil {
				outs = append(outs, err.Error())
				break
			}
			f, err2 := wr.Create(v)
			if err2 != nil {
				outs = append(outs, err2.Error())
				break
			}
			_, _ = f.Write(data)
		}
		wr.Close()

		if len(outs) > 0 {
			http.Error(w, strings.Join(outs, ";"), http.StatusInternalServerError)
			return
		}
		date := time.Now().Format(time.RFC3339)
		w.Header().Set("Content-Disposition", "attachment; filename=acc_"+date+".zip")
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	})

	mux.HandleFunc("/cgi-bin/ip_addr.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cmd := exec.Command("python3", "./cgi-bin/ip_addr.py")
		out, err := cmd.Output()
		if err != nil {
			http.Error(w, "could not obtain ip address", http.StatusInternalServerError)
			return
		}
		s := strings.ReplaceAll(string(out), "Content-type: text/json\n\n", "")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	})

	mux.HandleFunc("/cgi-bin/fileslist.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cmd := exec.Command("python3", "./cgi-bin/fileslist.py")
		out, err := cmd.Output()
		if err != nil {
			http.Error(w, "file acquisition error", http.StatusInternalServerError)
			return
		}
		s := strings.ReplaceAll(string(out), "Content-type: text/json\n\n", "")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	})

	mux.HandleFunc("/cgi-bin/dotest.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := "./cgi-bin/dotest.py"
		if os.Getenv("PSC_TRIAL") != "" {
			filename = "./cgi-bin/jim_dotest.py"
		}
		var d Indata
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		st, _ := json.Marshal(&d)
		cmd := exec.Command("python3", filename)
		cmd.Stdin = bytes.NewReader(st)
		out, err := cmd.Output()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s := strings.ReplaceAll(string(out), "Content-type: text/plain\n", "")
		s = strings.ReplaceAll(s, "Content-type: application/json\n", "")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	})

	mux.HandleFunc("/cgi-bin/initcontrolfile.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cmd := exec.Command("python3", "./cgi-bin/initcontrolfile.py")
		out, err := cmd.Output()
		if err != nil {
			http.Error(w, "could not initialize file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(out)
	})

	mux.HandleFunc("/cgi-bin/getdate.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s := time.Now().Format("2006-01-02T15:04:05-0700")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"datetime": s})
	})

	mux.HandleFunc("/cgi-bin/settime.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var data DateList
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		_ = SetSystemDate(makeDate(data))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"resp": true})
	})

	mux.HandleFunc("/flutter_service_worker.js", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	port := "8080"
	if os.Getenv("PSC_PORT") != "" {
		port = os.Getenv("PSC_PORT")
	}

	fmt.Printf("Listening on :%s\n", port)
	http.ListenAndServe(":"+port, mux)
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

// makeDate converts the DateList dt into a time.Time.
// dt provides the date (and optional time) components used to construct the returned time.Time.
// Expected component order is year, month, day, followed optionally by hour, minute, second, nanosecond;
// missing components default to zero and out-of-range values are normalized according to time.Date semantics.
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

	month := time.Month(m)

	return time.Date(y, month, d, hr, min, sec, msec, time.UTC)
}

func assureControlFile() {
	os.WriteFile("control.json", []byte("{}"), 0644)
}
