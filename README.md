# dumb-proxy-go

OOH OOH AHH AHH! 🦍🦍🦍 THIS IS DUMB-PROXY-GO!!! IT SUCKS SOCKS5 TRAFFIC IN, WRAPS IT IN A SINGLE MASTER WEBSOCKET CONNECTION WITH YAMUX MULTIPLEXING, AND BLASTS IT OUT TO THE INTERNET FAST!!! NO PACKET PARSING!!! ZERO BLOAT!!! FULL GORILLA EFFICIENCY!!! 🔥🔥🔥

## HOW COMPUTER DO DATA

[ BROWSER/OS ] --(MANY SOCKS5 LINES)--> [ DUMB PROXY CLIENT ] --(WEBSOCKET & YAMUX MULTIPLEX MULTIPLEX MULTIPLEX!)--> [ DUMB PROXY SERVER ] ----> [ REAL INTERNET!! ]

## HOW TO BUILD

EVERYTHING IS MERGED INTO ONE MONSTER MONOREPO!!! JUST RUN THE BUILD COMMAND ENGINE:

```bash
# Compile both client and server binaries at once!
make

# Or compile ONLY the proxy client!
make client

# Or compile ONLY the proxy server!
make server

# Clean up binaries when done
make clean
```
*Your compiled binaries will magically drop into the `bin/` directory!*

## HOW TO RUN FOR GORILLA STEPS

NO ARGUMENTS. NO CONFIG FLAGS. PORT :8080 AND :1080 ARE HARDCODED!!! ⚡⚡⚡

### 1. START THE SERVER (RUN ON THE CLOUD COMPUTER 🌋)
```bash
./bin/dumb-proxy-server
```
*Server runs on port `:8080` waiting for the master line!*

### 2. START THE CLIENT (RUN ON YOUR LOCAL LAPTOP 💻🦍)
```bash
./bin/dumb-proxy-client
```
*Client boots up, hooks the master WebSocket, and stands guard locally!*

### 3. USE SOCKS5 PROXY
Set your OS or browser proxy options to **SOCKS5 Proxy** pointing to `127.0.0.1:1080` and have internet access!!! 🍌🍌🍌

You can also test it instantly via terminal using curl:
```bash
curl -v --socks5-hostname 127.0.0.1:1080 https://google.com
```
