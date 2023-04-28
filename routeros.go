package main

type RouterOSHeader struct {
	title string
	path  string
	mac   bool
	copy  bool
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
				{"Disabled", "disabled", false, false},
				{"Inactive", "inactive", false, false},
				{"State", "current-state", false, false},
				{"Name", "name", false, false},
				{"Channel", "current-channel", false, false},
				{"Current Authorized Clients", "current-authorized-clients", false, false},
				{"L2 MTU", "l2mtu", false, false},
				{"Radio MAC", "radio-mac", false, false},
				{"Radio Name", "radio-name", false, false},
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
				{"Name", "name", false, false},
				{"Mode", "mode", false, false},
				{"SSID", "ssid", false, false},
				{"TX Chains", "tx-chains", false, false},
				{"RX Chains", "rx-chains", false, false},
				{"Country", "country", false, false},
				{"Installation", "installation", false, false},
				{"Security", "security", false, false},
				{"Datapath", "datapath", false, false},
				{"Channel", "channel", false, false},
			},
		},
		{
			title: "Channel",
			path:  "/caps-man/channel",
			headers: []RouterOSHeader{
				{"Name", "name", false, false},
				{"Frequency", "frequency", false, false},
				{"Control Channel Width", "control-channel-width", false, false},
				{"TX Power", "tx-power", false, false},
			},
		},
		{
			title: "Datapath",
			path:  "/caps-man/datapath",
			headers: []RouterOSHeader{
				{"Name", "name", false, false},
				{"Client to Client forwarding", "client-to-client-forwarding", false, false},
				{"Bridge", "bridge", false, false},
				{"Local forwarding", "local-forwarding", false, false},
			},
		},
		{
			title: "Security configuration",
			path:  "/caps-man/security",
			headers: []RouterOSHeader{
				{"Name", "name", false, false},
				{"Authentication Types", "authentication-types", false, false},
				{"Encryption", "encryption", false, false},
				{"Group Encryption", "group-encryption", false, false},
				{"Group Key Update", "group-key-update", false, false},
			},
		},
		{
			title: "Access List",
			path:  "/caps-man/access-list",
			headers: []RouterOSHeader{
				{"Interface", "interface", false, false},
				{"Signal Range", "signal-range", false, false},
				{"Client To Client Forwarding", "client-to-client-forwarding", false, false},
			},
		},
		{
			title: "Remote Cap",
			path:  "/caps-man/remote-cap",
			headers: []RouterOSHeader{
				{"Address", "address", true, false},
				{"Name", "name", false, false},
				{"Board", "board", false, false},
				{"Serial", "serial", false, false},
				{"Version", "version", false, false},
				{"Identity", "identity", false, false},
				{"Base Mac", "base-mac", true, false},
				{"State", "state", false, false},
				{"Radios", "radios", false, false},
			},
		},
		{
			title: "Radio",
			path:  "/caps-man/radio",
			headers: []RouterOSHeader{
				{"Radio Max", "radio-mac", false, false},
				{"Remote Cap Name", "remote-cap-name", false, false},
				{"Remote Cap Identity", "remote-cap-identity", false, false},
				{"Interface", "interface", false, false},
			},
		},
		{
			title: "Registration Table",
			path:  "/caps-man/registration-table",
			headers: []RouterOSHeader{
				{"Interface", "interface", false, false},
				{"SSID", "ssid", false, false},
				{"Mac-Address", "mac-address", true, false},
				{"EAP Identity", "eap-identity", false, false},
				{"Tx Rate", "tx-rate", false, false},
				{"Tx signal", "tx-rate-set", false, false},
				{"Rx Rate", "rx-rate", false, false},
				{"Rx signal", "rx-signal", false, false},
				{"Uptime", "uptime", false, false},
				{"Tx/Rx Packets", "packets", false, false},
				{"Tx/Rx Bytes", "bytes", false, false},
			},
		},
	},
	"Interfaces": {
		{
			title: "Interface",
			path:  "/interface/ethernet",
			headers: []RouterOSHeader{
				{"Name", "name", false, false},
				{"Actual MTU", "mtu", false, false},
				{"L2 MTU", "l2mtu", false, false},
				{"TX", "tx-bytes", false, false},
				{"RX", "rx-bytes", false, false},
			},
		},
	},
	"Wireless": {
		{
			title: "WiFi Interfaces",
			path:  "/interface/wireless",
			headers: []RouterOSHeader{
				{"Name", "name", false, false},
				{"Actual MTU", "mtu", false, false},
				{"MAC Address", "mac-address", true, false},
				{"ARP", "arp", false, false},
				{"Mode", "mode", false, false},
				{"Band", "band", false, false},
				{"Channel Width", "channel-width", false, false},
				{"Frequency", "frequency", false, false},
				{"SSID", "ssid", false, false},
			},
		},
	},
	"Bridge": {
		{
			title: "Host",
			path:  "/interface/bridge/host",
			headers: []RouterOSHeader{
				{"MAC Address", "mac-address", true, false},
				{"On Interface", "on-interface", false, false},
				{"Bridge", "bridge", false, false},
			},
		},
	},
	"ARP": {
		{
			title: "ARP Table",
			path:  "/ip/arp",
			headers: []RouterOSHeader{
				{"IP Address", "address", false, true},
				{"MAC Address", "mac-address", true, false},
				{"Interface", "interface", false, false},
			},
		},
	},
	"DHCP Server": {
		{
			title: "Leases",
			path:  "/ip/dhcp-server/lease",
			headers: []RouterOSHeader{
				{"Address", "address", false, true},
				{"MAC Address", "mac-address", true, false},
				{"Client ID", "active-client-id", false, false},
				{"Server", "server", false, false},
				{"Active Address", "active-address", false, true},
				{"Active MAC Address", "active-mac-address", true, false},
				{"Host Name", "host-name", false, false},
				{"Expires After", "expires-after", false, false},
			},
		},
	},
}
