interfaces {
    ethernet eth0 {
        address dhcp
        description UPLINK
        duplex auto
        firewall {
            in {
                name WAN-IN
            }
        }
        speed auto
    }
    ethernet eth1 {
        address 192.168.0.1/24
        description HOUSE
        duplex auto
        speed auto
    }
    loopback lo {
    }
    switch switch0 {
        mtu 1500
    }
}
