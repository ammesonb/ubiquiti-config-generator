firewall {
    all-ping enable
    broadcast-ping disable
    group {
        address-group admin {
            /* Reserved hosts for admin stuff */
            description admin
            address 192.168.0.1
            address 192.168.0.2
            address 192.168.1.1
        }
    }
    log-martians enable
    name WAN_IN {
        default-action drop
        rule 100 {
            action accept
            description "Allow 'IGMP'"
            log disable
            protocol igmp
        }
    }
}
