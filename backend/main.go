package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var filesloc = ""
var archives = ""

const logPath = "/dev/shm/control.json"

var (
	currentCmd *exec.Cmd
	mu         sync.Mutex
)

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

func getUser() string {
	j, _ := user.Current()
	if j != nil {
		return j.Username
	}
	return "unknown"
}
func main() {

	// log.SetOutput(os.Stdout)
	assureControlFile()
	makeReportDirs()
	user := getUser()
	fmt.Printf("Running server as user: %s\n", user)

	mux := http.NewServeMux()

	mime.AddExtensionType(".dart.js", "text/javascript")
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".css", "text/css")

	wwwroot := "."

	_, err := os.Stat("index.html")
	if errors.Is(err, os.ErrNotExist) {
		wwwroot = "./build"
	}

	// _, err := os.Stat("build")
	// if errors.Is(err, os.ErrNotExist) {
	// 	wwwroot = ""
	// }
	// server side events for the web app

	mux.HandleFunc("/cgi-bin/dotest.py", startHandler)
	mux.HandleFunc("/stop", stopHandler)
	mux.HandleFunc("/events", sseHandler)

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

	mux.HandleFunc("/cgi-bin/get_combined.py", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		filesRaw := r.FormValue("files")
		files := strings.Split(filesRaw, ",")

		// Filename: combined_logs_2026-04-25_17-00.csv
		date := time.Now().Format("2006-01-02_15-04")

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=\"combined_logs_"+date+".csv\"")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Use a buffered writer for efficiency
		bw := bufio.NewWriter(w)
		defer bw.Flush()

		for _, v := range files {
			if v == "" {
				continue
			}

			safeName := filepath.Base(v)
			fullPath := filepath.Join(filesloc, safeName)

			file, err := os.Open(fullPath)
			if err != nil {
				// Log to terminal so you can see if a file was missing
				fmt.Fprintf(os.Stderr, "Warning: could not find %s\n", fullPath)
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				// Prepend filename as the first column
				line := scanner.Text()
				if line != "" {
					fmt.Fprintf(bw, "%s,%s\n", safeName, line)
				}
			}
			file.Close()
		}
	})

	mux.HandleFunc("/cgi-bin/get_zip.py", func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("Zip requested with method: %s\n", r.Method)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// if err := r.ParseForm(); err != nil {
		// 	http.Error(w, "invalid form", http.StatusBadRequest)
		// 	return
		// }

		selectedFiles := strings.Split(r.FormValue("files"), ",")

		fmt.Printf("Requested files for zip: %v\n", selectedFiles)
		if len(selectedFiles) == 0 {
			http.Error(w, "no files selected", http.StatusBadRequest)
			return
		}

		// 1. Filename prep: avoid colons (illegal in many OS filenames)
		// Using a simpler format: acc_2026-04-25_13-00.zip
		date := time.Now().Format("2006-01-02_15-04")

		// 2. Set headers BEFORE writing any data
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=acc_"+date+".csv")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 3. Stream directly to ResponseWriter to save Pi memory
		zipWriter := zip.NewWriter(w)

		// Finalize the zip at the end of the function
		defer zipWriter.Close()

		for _, fileName := range selectedFiles {
			// Use filepath.Base to prevent "Directory Traversal" attacks
			// (someone trying to request ../../../etc/passwd)
			safeName := filepath.Base(fileName)
			fullPath := filepath.Join(filesloc, safeName)

			file, err := os.Open(fullPath)
			if err != nil {
				// We can't stop the HTTP stream now, so we skip the missing file
				continue
			}

			// Create the entry in the zip using the original filename
			entryWriter, err := zipWriter.Create(safeName)
			if err != nil {
				file.Close()
				continue
			}

			// 4. Efficiently stream from SD card to Network
			_, _ = io.Copy(entryWriter, file)
			file.Close()
		}
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

	// mux.HandleFunc("/cgi-bin/dotest.py", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.Method != http.MethodPost {
	// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	// 		return
	// 	}
	// 	filename := "./cgi-bin/dotest.py"
	// 	if user != "root" {
	// 		filename = "./cgi-bin/jim_dotest.py"
	// 	}
	// 	var d Indata
	// 	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
	// 		http.Error(w, "invalid json", http.StatusBadRequest)
	// 		return
	// 	}
	// 	st, _ := json.Marshal(&d)
	// 	cmd := exec.Command("python3", filename)
	// 	cmd.Stdin = bytes.NewReader(st)
	// 	out, err := cmd.Output()
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		fmt.Printf("$err\n")
	// 		return
	// 	}
	// 	s := strings.ReplaceAll(string(out), "Content-type: text/plain\n", "")
	// 	s = strings.ReplaceAll(s, "Content-type: application/json\n", "")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(s))
	// })

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
	// redirect all http traffic to https
	go func() {
		http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://psctester.local"+r.RequestURI, http.StatusMovedPermanently)
		}))
	}()

	port := "8080"
	certPath := "/etc/ssl/certs/psctester.cert.pem"
	keyPath := "/etc/ssl/private/psctester.key.pem"
	if user == "root" {
		port = "443"
	} else if os.Getenv("PSC_PORT") != "" {
		port = os.Getenv("PSC_PORT")
	} else {
		port = "8080"
	}

	fmt.Printf("Listening on :%s\n", port)
	if port == "443" {
		fmt.Printf("Using TLS with cert: %s and key: %s\n", certPath, keyPath)
		err := http.ListenAndServeTLS(":"+port, certPath, keyPath, mux)
		if err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		http.ListenAndServe(":"+port, mux)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// 2. Setup fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Watcher error: %v", err)
		return
	}
	defer watcher.Close()

	// Monitor the specific control file
	controlFilePath := "/dev/shm/control.json"
	err = watcher.Add(controlFilePath)
	if err != nil {
		log.Printf("Add file error: %v", err)
	}

	// 3. Setup Heartbeat Ticker (15 seconds)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	fmt.Println("Client connected to SSE")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only trigger on Write events (or Chmod if metadata changes)
			if event.Op&fsnotify.Write == fsnotify.Write {
				data, err := os.ReadFile(controlFilePath)
				if err == nil && len(data) > 0 {

					var js json.RawMessage
					if err := json.Unmarshal(data, &js); err == nil {

						fmt.Fprintf(w, "data: %s\n\n", string(data))
						flusher.Flush()
					} else {
						log.Printf("Invalid JSON in control file: %s", string(data))
					}
				}
			}

		case <-ticker.C:
			// Send heartbeat to keep the connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			// Client closed the connection
			fmt.Println("Client disconnected")
			return
		}
	}
}
func sendUpdate(w http.ResponseWriter, flusher http.Flusher) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "%s\n\n", string(data))
	flusher.Flush()
}
func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST", 405)
		return
	}

	jsondata, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read error", 400)
		return
	}

	mu.Lock()
	// Don't start if already running
	if currentCmd != nil && (currentCmd.Process != nil && currentCmd.ProcessState == nil) {
		mu.Unlock()
		http.Error(w, "Process already running", 409)
		return
	}

	filename := "./cgi-bin/dotest.py"
	if os.Getenv("TESTING") == "1" {
		filename = "./cgi-bin/jim_dotest.py"
	}

	cmd := exec.Command("/usr/bin/python3", "-u", filename)

	// CRITICAL: Connect Python's output to Go's console so you can see prints/errors
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		mu.Unlock()
		http.Error(w, "Pipe error", 500)
		return
	}

	if err := cmd.Start(); err != nil {
		mu.Unlock()
		http.Error(w, "Failed to start", 500)
		currentCmd = nil
		return
	}

	currentCmd = cmd
	mu.Unlock() // Unlock early so other handlers can check currentCmd

	// Handle Stdin and Waiting in the background
	go func() {
		// This 'defer' tells Python: "That's all the data you're getting!"
		defer stdin.Close()

		_, err := stdin.Write(jsondata)
		if err != nil {
			fmt.Printf("Error writing to Python stdin: %v\n", err)
			return
		}
		// fmt.Printf("Successfully wrote %d bytes to Python stdin\n", n)
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "started", "pid": %d}`, cmd.Process.Pid)
	go func(cmd *exec.Cmd) {
		err := cmd.Wait() // This populates ProcessState

		mu.Lock()

		if err != nil {
			log.Printf("Process exited with error: %v", err)
		} else {
			log.Println("Process finished successfully")
		}
		currentCmd = nil
		mu.Unlock()
		// (Depending on if you want to keep the record of the last run)
	}(currentCmd)

	fmt.Fprint(w, "Started")

}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if currentCmd != nil && currentCmd.Process != nil {
		currentCmd.Process.Signal(os.Interrupt)
		w.WriteHeader(http.StatusOK)
	}
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
