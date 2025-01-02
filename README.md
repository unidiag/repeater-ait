# repeater-ait

This program retransmits incoming UDP to the next port with the addition of the AIT.

## compile
`go build -ldflags "-linkmode external -extldflags '-static'" -o repeater-ait`

## usage
`./repeater-ait <udp:port> <hbblink>`

## example
`./repeater-ait udp://eth1@239.0.100.1:1234 http://hbbtv.com/app`


![Screenshot cascap](https://github.com/unidiag/repeater-ait/blob/main/screenshot.jpg)