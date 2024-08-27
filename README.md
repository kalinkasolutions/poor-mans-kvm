# Poor man's kvm

This tool is designed for USB hubs with two input sources. It sends a command to all connected screens when a specific device disconnects from the computer where the tool is running.

For example, if a USB hub is connected to both a computer and a laptop, and pressing the hub's button causes the keyboard to disconnect, the tool will switch the monitor's input source to the laptop.

If the devices reconnects the tool will switch the monitor's input source back.

### Create the config
#### Find display buses
<pre>
<b>$ ddcutil detect</b>

[...]
Display 1
   I2C bus:  /dev/i2c-6 <b><-- DisplayBusNumber = 6</b>
   EDID synopsis:
      Mfg id:               BNQ
      Model:                BenQ XL2420T
      Product code:         32518
      Serial number:        68C00272SL0
      Binary serial number: 21573 (0x00005445)
      Manufacture year:     2012,  Week: 33
   VCP version:         2.0
[...]
</pre>
Find out which code controls your input source (probably VCP code: 60)


#### Find input switch code
<pre>
<b>$ ddcutil vcpinfo --brief</b>

[...]
VCP code: 5E: 6 axis saturation: Magenta
VCP code: 60: Input Source <b><-- VcpInputSourceCode = 60</b>
VCP code: 62: Audio speaker volume
[...]
</pre>

#### Find display input codes
<pre>
<b>$ ddcutil --dis 1 cap --verbose</b>

[...]
Feature: 52 (Active control)
   Feature: 60 (Input Source)
      Values (unparsed): 01 03 11 12 0F
      Values (  parsed):
         01: VGA-1 
         03: DVI-1
         11: HDMI-1 <b><-- e,g DisconnectInputCode = 0x11</b>
         12: HDMI-2
         0f: DisplayPort-1 <b><-- e.g ConnectInputCode = 0x0f</b>
   Feature: 62 (Audio speaker volume)
[...]
</pre>

#### Find usb device to listen to
<pre>
<b>lsusb</b>

[...]
Bus 001 Device 032: ID 046a:c098 Cherry GmbH CHERRY Corded Device
Bus 001 Device 031: ID 046d:c539 Logitech, Inc. USB Receiver <b><-- DeviceID which connects and disconnects</b>
Bus 001 Device 030: ID 05e3:0610 Genesys Logic, Inc. Hub
[...]
</pre>

#### example config
<pre>
<b>config.json</b>

{
  "DeviceID": "Cherry GmbH CHERRY Corded Device",
  "ConnectInputCode": "0x0f",
  "DisconnectInputCode": "0x12",
  "DisplayBusNumbers": ["6", "8"],
  "UsbPollingIntervalMs": 500,
  "VcpInputSourceCode": "60"
}
</pre>

### Create service: 

```
  go build
$ mv poor-mans-kvm /usr/local/bin/poor-mans-kvm
```

create file

<pre>
<b>/etc/systemd/poormanskvm.service</b>

[Unit]
Description=Poor Man's KVM Service
After=network.target

[Service]
ExecStart=/usr/local/bin/poor-mans-kvm -configLocation="/path/to/config.json"
Restart=on-failure

[Install]
WantedBy=multi-user.target
</pre>

### Start the service
```
sudo systemctl daemon-reload
sudo systemctl enable poormanskvm
sudo systemctl start poormanskvm
```