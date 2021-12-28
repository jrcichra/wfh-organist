# wfh-church-organist

Be a Work-From-Home Church Organist. Send regular MIDI over regular TCP/IP.

# Introduction

This program listens to MIDI input and sends the notes over TCP. The program used in server or client mode, or both at the same time. This leads to some interesting use cases:

+ Remote control a MIDI keyboard over the LAN
+ Remote control a MIDI keyboard over the Internet
+ Test modifications to `server.go` or `client.go` locally knowning it would work the same on the LAN or the Internet, because the program goes through the IP stack regardless of what mode
+ Conditionally modify MIDI channels in `client.go` to work with the organ attached on the other end

# Build notes
I used Go 1.17 for this project, but older versions will probably work. There are external cgo dependencies so you'll need a few packages from your distro's package manager. This also means I can't easily provide cross-architecture targets

# Usage
+ Download a recent version of [Go](https://go.dev/dl/) for your operating system 
+ `git clone https://github.com/jrcichra/wfh-church-organist.git`
+ `go build`
+ `./wfh-church-organist -help`

```
Usage of ./wfh-church-organist:
  -list
        list available ports
  -midi int
        midi port (default 0)
  -mode string
        client, server, or local (runs both) (default "local")
  -port int
        server port (default 3131)
  -server string
        server IP (default "localhost")
```

# Design choices
+ Simplicity - It should be easy to understand what the code is doing
+ TCP - This program was implimented with TCP but could also use UDP. I chose TCP to avoid 'stuck notes' in the event a NoteOff packet was dropped. TCP has the downside of effectively 'losing notes'. When a lag spike hits, the TCP stream will catch up and all the MIDI events will happen as fast as possible, which does not respect the original timing.
+ Single Binary - Instead of managing two binaries, "server/client", I combined them into a single binary. I felt the space increase was worth the flexability and simplicity of managing one binary. The mode is controlled with a single flag. There was also a lot of shared code between the server and client, so making it a single binary was easy.

# Disclaimer
This program is not intended for production use. I do not claim that this will work flawlessly for remote performances. Please anaylze the code and determine if your connection stability and latency will work with the way I have implimented this program.

# Testing
This program was tested under a variety of real-world latency conditions, all with minimal packet loss. Your mileage may vary.
+ Local mode - `0-1ms` delay
+ Starlink -> Linode Canada  -> Starlink - `50-70ms` delay  - `Great experience`
+ Starlink -> Oracle Ashburn -> Starlink - `30-40ms` delay  - `Good experience`
+ T-Mobile IoT -> Oracle Ashburn -> T-Mobile IoT - `100-170ms` delay - `Tolerable experience`