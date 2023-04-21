package main

type RouterOSHeader struct {
	title string
	path  string
}

type RouterOSView struct {
	title   string
	path    string
	headers []RouterOSHeader
}

var routerOStree = map[string][]string{
	"":       {"CAPsMAN", "Wireless", "Interfaces", "IP", "System"},
	"IP":     {"ARP", "DHCP Server"},
	"System": {"Certificates", "Health"},
}

var routerOSCommands = map[string][]RouterOSView{
	"CAPsMAN": {
		{
			title: "Interfaces",
			path:  "/caps-man/interface",
			headers: []RouterOSHeader{
				{"Disabled", "disabled"},
				{"Inactive", "inactive"},
				{"State", "current-state"},
				{"Name", "name"},
				{"Channel", "current-channel"},
				{"Current Authorized Clients", "current-authorized-clients"},
				{"L2 MTU", "l2mtu"},
				{"Radio MAC", "radio-mac"},
				{"Radio Name", "radio-name"},
			},
		},
		{
			title:   "Provisioning",
			path:    "/caps-man/provisioning",
			headers: []RouterOSHeader{},
		},
		{
			title: "Configuration",
			path:  "/caps-man/configuration",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Mode", "mode"},
				{"SSID", "ssid"},
				{"TX Chains", "tx-chains"},
				{"RX Chains", "rx-chains"},
				{"Country", "country"},
				{"Installation", "installation"},
				{"Security", "security"},
				{"Datapath", "datapath"},
				{"Channel", "channel"},
			},
		},
		{
			title: "Channel",
			path:  "/caps-man/channel",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Frequency", "frequency"},
				{"Control Channel Width", "control-channel-width"},
				{"TX Power", "tx-power"},
			},
		},
		{
			title: "Datapath",
			path:  "/caps-man/datapath",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Client to Client forwarding", "client-to-client-forwarding"},
				{"Bridge", "bridge"},
				{"Local forwarding", "local-forwarding"},
			},
		},
		{
			title: "Security configuration",
			path:  "/caps-man/security",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Authentication Types", "authentication-types"},
				{"Encryption", "encryption"},
				{"Group Encryption", "group-encryption"},
				{"Group Key Update", "group-key-update"},
			},
		},
		{
			title: "Access List",
			path:  "/caps-man/access-list",
			headers: []RouterOSHeader{
				{"Interface", "interface"},
				{"Signal Range", "signal-range"},
				{"Client To Client Forwarding", "client-to-client-forwarding"},
			},
		},
		{
			title: "Remote Cap",
			path:  "/caps-man/remote-cap",
			headers: []RouterOSHeader{
				{"Address", "address"},
				{"Name", "name"},
				{"Board", "board"},
				{"Serial", "serial"},
				{"Version", "version"},
				{"Identity", "identity"},
				{"Base Mac", "base-mac"},
				{"State", "state"},
				{"Radios", "radios"},
			},
		},
		{
			title: "Radio",
			path:  "/caps-man/radio",
			headers: []RouterOSHeader{
				{"Radio Max", "radio-mac"},
				{"Remote Cap Name", "remote-cap-name"},
				{"Remote Cap Identity", "remote-cap-identity"},
				{"Interface", "interface"},
			},
		},
		{
			title: "Registration Table",
			path:  "/caps-man/registration-table",
			headers: []RouterOSHeader{
				{"Interface", "interface"},
				{"SSID", "ssid"},
				{"Mac-Address", "mac-address"},
				{"EAP Identity", "eap-identity"},
				{"Tx Rate", "tx-rate"},
				{"Tx signal", "tx-rate-set"},
				{"Rx Rate", "rx-rate"},
				{"Rx signal", "rx-signal"},
				{"Uptime", "uptime"},
				{"Tx/Rx Packets", "packets"},
				{"Tx/Rx Bytes", "bytes"},
			},
		},
	},
	"Interfaces": {
		{
			title: "Interface",
			path:  "/interface/ethernet",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Actual MTU", "mtu"},
				{"L2 MTU", "l2mtu"},
				{"TX", "tx-bytes"},
				{"RX", "rx-bytes"},
			},
		},
	},
	"Wireless": {
		{
			title: "WiFi Interfaces",
			path:  "/interface/wireless",
			headers: []RouterOSHeader{
				{"Name", "name"},
				{"Actual MTU", "mtu"},
				{"MAC Address", "mac-address"},
				{"ARP", "arp"},
				{"Mode", "mode"},
				{"Band", "band"},
				{"Channel Width", "channel-width"},
				{"Frequency", "frequency"},
				{"SSID", "ssid"},
			},
		},
	},
	"ARP": {
		{
			title: "ARP Table",
			path:  "/ip/arp",
			headers: []RouterOSHeader{
				{"IP Address", "address"},
				{"MAC Address", "mac-address"},
				{"Interface", "interface"},
			},
		},
	},
	"DHCP Server": {
		{
			title: "Leases",
			path:  "/ip/dhcp-server/lease",
			headers: []RouterOSHeader{
				{"Disabled", "disabled"},
				{"Dynamic", "dynamic"},
				{"Status", "status"},
				{"Address", "address"},
				{"MAC Address", "mac-address"},
				{"Client ID", "active-client-id"},
				{"Server", "server"},
				{"Active Address", "active-address"},
				{"Active MAC Address", "active-mac-address"},
				{"Host Name", "host-name"},
				{"Expires After", "expires-after"},
			},
		},
	},
}
