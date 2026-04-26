Just unpack this as a folder under /home/pi

You may run 

sudo setcap 'cap_net_bind_service=+ep' /path/to/your/compiled_go_binary

or run as root or sudo, it runs on port 443 (80 redirects to 443), and actual connections and testing and encryption are enabled.

cgi python scripts are in the psctester/cgi-bin folder

For the developer:

If you want to recompile the binary, psctester, do the following.

wget https://dl.google.com/go/go1.18.8.linux-armv6l.tar.gz -O go.tar.gz
(this is the latest version available as of November, 2022)
later versions of go will probably work fine

sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xvf go.tar.gz
go version

if you change anything in the web folder, the backend folder, or the web/web folder run 
    $ make.sh
from the main folder (with pubspec.yaml)

you might need to do
    dart pub global activate webdev
When you run build, the stuff in the web folder gets compiled to javascript and 
goes into the build folder. 

Once everything is compiled, the build folder can be used independently as the complete application.

It is ok to run 

sudo setcap 'cap_net_bind_service=+ep' build/psctester

to run the app as a user other than root.


