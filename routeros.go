package main

type RouterOSHeader struct {
	title string
	path  string
	mac   bool
}

type RouterOSView struct {
	title   string
	path    string
	headers []RouterOSHeader
}

var routerOStree = map[string][]string{
	"":       {"CAPsMAN", "Wireless", "Interfaces", "Bridge", "IP", "System"},
	"IP":     {"ARP", "DHCP Server"},
	"System": {"Certificates", "Health"},
}

var routerOSCommands = map[string][]RouterOSView{
	"CAPsMAN": {
		{
			title: "Interfaces",
			path:  "/caps-man/interface",
			headers: []RouterOSHeader{
				{"Disabled", "disabled", false},
				{"Inactive", "inactive", false},
				{"State", "current-state", false},
				{"Name", "name", false},
				{"Channel", "current-channel", false},
				{"Current Authorized Clients", "current-authorized-clients", false},
				{"L2 MTU", "l2mtu", false},
				{"Radio MAC", "radio-mac", false},
				{"Radio Name", "radio-name", false},
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
				{"Name", "name", false},
				{"Mode", "mode", false},
				{"SSID", "ssid", false},
				{"TX Chains", "tx-chains", false},
				{"RX Chains", "rx-chains", false},
				{"Country", "country", false},
				{"Installation", "installation", false},
				{"Security", "security", false},
				{"Datapath", "datapath", false},
				{"Channel", "channel", false},
			},
		},
		{
			title: "Channel",
			path:  "/caps-man/channel",
			headers: []RouterOSHeader{
				{"Name", "name", false},
				{"Frequency", "frequency", false},
				{"Control Channel Width", "control-channel-width", false},
				{"TX Power", "tx-power", false},
			},
		},
		{
			title: "Datapath",
			path:  "/caps-man/datapath",
			headers: []RouterOSHeader{
				{"Name", "name", false},
				{"Client to Client forwarding", "client-to-client-forwarding", false},
				{"Bridge", "bridge", false},
				{"Local forwarding", "local-forwarding", false},
			},
		},
		{
			title: "Security configuration",
			path:  "/caps-man/security",
			headers: []RouterOSHeader{
				{"Name", "name", false},
				{"Authentication Types", "authentication-types", false},
				{"Encryption", "encryption", false},
				{"Group Encryption", "group-encryption", false},
				{"Group Key Update", "group-key-update", false},
			},
		},
		{
			title: "Access List",
			path:  "/caps-man/access-list",
			headers: []RouterOSHeader{
				{"Interface", "interface", false},
				{"Signal Range", "signal-range", false},
				{"Client To Client Forwarding", "client-to-client-forwarding", false},
			},
		},
		{
			title: "Remote Cap",
			path:  "/caps-man/remote-cap",
			headers: []RouterOSHeader{
				{"Address", "address", true},
				{"Name", "name", false},
				{"Board", "board", false},
				{"Serial", "serial", false},
				{"Version", "version", false},
				{"Identity", "identity", false},
				{"Base Mac", "base-mac", true},
				{"State", "state", false},
				{"Radios", "radios", false},
			},
		},
		{
			title: "Radio",
			path:  "/caps-man/radio",
			headers: []RouterOSHeader{
				{"Radio Max", "radio-mac", false},
				{"Remote Cap Name", "remote-cap-name", false},
				{"Remote Cap Identity", "remote-cap-identity", false},
				{"Interface", "interface", false},
			},
		},
		{
			title: "Registration Table",
			path:  "/caps-man/registration-table",
			headers: []RouterOSHeader{
				{"Interface", "interface", false},
				{"SSID", "ssid", false},
				{"Mac-Address", "mac-address", true},
				{"EAP Identity", "eap-identity", false},
				{"Tx Rate", "tx-rate", false},
				{"Tx signal", "tx-rate-set", false},
				{"Rx Rate", "rx-rate", false},
				{"Rx signal", "rx-signal", false},
				{"Uptime", "uptime", false},
				{"Tx/Rx Packets", "packets", false},
				{"Tx/Rx Bytes", "bytes", false},
			},
		},
	},
	"Interfaces": {
		{
			title: "Interface",
			path:  "/interface/ethernet",
			headers: []RouterOSHeader{
				{"Name", "name", false},
				{"Actual MTU", "mtu", false},
				{"L2 MTU", "l2mtu", false},
				{"TX", "tx-bytes", false},
				{"RX", "rx-bytes", false},
			},
		},
	},
	"Wireless": {
		{
			title: "WiFi Interfaces",
			path:  "/interface/wireless",
			headers: []RouterOSHeader{
				{"Name", "name", false},
				{"Actual MTU", "mtu", false},
				{"MAC Address", "mac-address", true},
				{"ARP", "arp", false},
				{"Mode", "mode", false},
				{"Band", "band", false},
				{"Channel Width", "channel-width", false},
				{"Frequency", "frequency", false},
				{"SSID", "ssid", false},
			},
		},
	},
	"Bridge": {
		{
			title: "Host",
			path:  "/interface/bridge/host",
			headers: []RouterOSHeader{
				{"MAC Address", "mac-address", true},
				{"On Interface", "on-interface", false},
				{"Bridge", "bridge", false},
			},
		},
	},
	"ARP": {
		{
			title: "ARP Table",
			path:  "/ip/arp",
			headers: []RouterOSHeader{
				{"IP Address", "address", false},
				{"MAC Address", "mac-address", true},
				{"Interface", "interface", false},
			},
		},
	},
	"DHCP Server": {
		{
			title: "Leases",
			path:  "/ip/dhcp-server/lease",
			headers: []RouterOSHeader{
				{"Address", "address", false},
				{"MAC Address", "mac-address", true},
				{"Client ID", "active-client-id", false},
				{"Server", "server", false},
				{"Active Address", "active-address", false},
				{"Active MAC Address", "active-mac-address", true},
				{"Host Name", "host-name", false},
				{"Expires After", "expires-after", false},
			},
		},
	},
}
