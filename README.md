# wolproxy

A Wake-on-LAN proxy. It listens for WoL magic packets on UDP and, when one
targets a specific MAC, emits a raw Ethernet magic packet for a *different* MAC
out a chosen interface.

## Why

It turns out that [Steam Link](https://store.steampowered.com/remoteplay) can
send WoL packets to any device it has been associated with. Unfortunately for
me, some of my devices are only connected over WiFi.

While I've tried to get WoL to work over WiFi, I've hit a number of problems
making that a dead end:
* my WiFi PCIe adapter is based on the AX200 Intel chipset which loses
  association during suspend-to-RAM, and comes back online un-associated,
* the motherboard on the target machine would fail to suspend correctly to
  Suspend-to-Idle due to a bug in firmware, which in turn would cause the GPU to
  not come back online on resume.

But I do have a second machine running (physically) near my desktop connected to
the same network over WiFi and with a spare ethernet adapter. I can easily
connect the two over Ethernet. And since WoL happens at L2 (or over UDP
broadcasts), I can have that second machine relay WoL packets from its wireless
interface to the wired one of my desktop. Steam Link is none the wiser of this
machinery existing, it just sends the WoL packet over broadcast and expects the
desktop to come online eventually. Magic :sparkles:.

## Configuration

All configuration is via environment variables:

| Variable | Meaning | Required |
|---|---|---|
| `WOLPROXY_MATCH_MAC` | The actual MAC address used as target of the WoL packet | yes |
| `WOLPROXY_TARGET_MAC` | The MAC of the interface to actually wake, reachable over `WOLPROXY_IFACE` | yes |
| `WOLPROXY_IFACE` | The interface to emit the frame from | yes |
| `WOLPROXY_LISTEN` | UDP listen address for inbound magic packets (default `:9`) | no |

See [`wolproxy.env.example`](wolproxy.env.example) for a template.

## Requirements

- Go 1.26+ to build.
- `CAP_NET_RAW` (raw packet socket) and, for the default port 9,
  `CAP_NET_BIND_SERVICE`. The systemd unit grants both; otherwise run as root.
- `WOLPROXY_IFACE` must be up with a non-zero MAC when the process starts — that
  MAC becomes the frame's source address.

## Build

Native build:

```sh
make build
```

Cross-compile for a 64-bit Linux from another machine:

```sh
make build-linux              # GOOS=linux GOARCH=arm64
make build-linux GOARCH=arm   # 32-bit ARM Linux
```

## Deploy

First create your config from the template:

```sh
cp wolproxy.env.example wolproxy.env
$EDITOR wolproxy.env                 # fill in the MACs and interface
```

Then, on the proxying machine:

```sh
sudo make install
sudo systemctl daemon-reload
sudo systemctl enable --now wolproxy
```

`make install` places:

- the binary at `/usr/local/bin/wolproxy`
- the unit at `/etc/systemd/system/wolproxy.service`
- `/etc/wolproxy.env` (0600) from `wolproxy.env`, if it does not already exist

If you cross-compiled on another machine, copy `wolproxy`, `wolproxy.service`,
and your `wolproxy.env` to the proxying machine, or run `make install` there
against the copied tree.

Paths are overridable: `make install PREFIX=/opt SYSTEMD_DIR=... ENV_FILE=...`.

### Verify

```sh
systemctl status wolproxy
journalctl -u wolproxy -f
```

Each received packet logs its source and parsed target MAC, and whether it
matched and was forwarded.

## Uninstall

```sh
sudo systemctl disable --now wolproxy
sudo make uninstall     # leaves /etc/wolproxy.env in place
```
