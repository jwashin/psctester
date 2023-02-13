Just unpack this as a folder under /home/pi

If you want to recompile the binary, psctester, do the following.

wget https://dl.google.com/go/go1.18.8.linux-armv6l.tar.gz -O go.tar.gz
(this is the latest version available as if November, 2022)
later versions of go will probably work fine

sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xvf go.tar.gz
go version

To build the executable, 
    $ go build
from the main folder with main.go and go.mod

If the executable really needs to run as root, suid the psctester binary.
sudo chown root:pi psctester
sudo chmod u+s psctester

psctester runs on port 8080 by default. set an environment variable PSC_PORT to
run on a different port.

export PSC_PORT=80

psctester runs in debug mode by default. set an environment variable GIN_MODE to
run in release mode.

export GIN_MODE=release

dart may be installed from the instructions at https://dart.dev/get-dart
but you only need it if you change the javascript in the web folder.

if you change anything in the web folder, run 
    $ webdev build
from the main folder (with pubspec.yaml)
you might need to do
    dart pub global activate webdev
When you run build, the stuff in the web folder gets compiled to javascript and 
goes into the build folder. psctester serves the user interface from there.

Python scripts stay in cgi-bin.

An env variable PSC_TRIAL, when present, is for evaluating the user interface without 
real things happening.

$ set PSC_TRIAL=1  /sets
$ set PSC_TRIAL=   /unsets



