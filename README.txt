Just unpack this as a folder under /home/pi

When run as pi or other users, psctester runs on port 8080, with only simulated data 

If run as root or sudo, it runs on port 80, and actual connections and testing are enabled.

cgi python scripts go in the psctester/build/cgi-bin folder

For the developer:

If you want to recompile the binary, psctester, do the following.

wget https://dl.google.com/go/go1.18.8.linux-armv6l.tar.gz -O go.tar.gz
(this is the latest version available as of November, 2022)
later versions of go will probably work fine

sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xvf go.tar.gz
go version

To build the executable, 
    $ go build
from the main folder with main.go and go.mod


dart may be installed from the instructions at https://dart.dev/get-dart
but you only need it if you change the javascript in the web folder.

if you change anything in the web folder, run 
    $ webdev build
from the main folder (with pubspec.yaml)

In a pinch, to compile the main javascript file from dart, 
    $ dart compile js psctester.dart -o psctester.dart.js

you might need to do
    dart pub global activate webdev
When you run build, the stuff in the web folder gets compiled to javascript and 
goes into the build folder. psctester serves the user interface from there.

Python scripts stay in cgi-bin.



