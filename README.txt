Unpack this as a folder under /home/pi/acc_tester. It should be called www.

The web app js, html, and css file are in this folder.

The included binary, psctester, was developed for testing, but it will
serve the web app just fine in this configuration.

If you want to recompile the binary, psctester, do the following to install 
go and compile the package.

wget https://dl.google.com/go/go1.18.8.linux-armv6l.tar.gz -O go.tar.gz
(this is the latest version available as if November, 2022)
later versions of go will probably work fine

sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xvf go.tar.gz
(assure /usr/local/go/bin is in your executable path)
go version

once Go is installed, you may run the app.
go main.go

To build the executable,
    $ go build
from the www folder with main.go and go.mod

If the executable really needs to run as root, suid the psctester binary.
sudo chown root:pi psctester
sudo chmod u+s psctester

psctester runs on port 8080 by default. set an environment variable PSC_PORT to
run on a different port.

export PSC_PORT=80

psctester runs in debug mode by default. set an environment variable GIN_MODE to
run in release mode.

export GIN_MODE=release

Python scripts stay in cgi-bin. Some are actually used by the go program by invoking them as
CGI using python3. Changes to python scripts should be in cgi-bin.

An env variable PSC_TRIAL, when present, is for evaluating the user interface without
real things happening.

$ set PSC_TRIAL=1  /sets
$ set PSC_TRIAL=   /unsets

