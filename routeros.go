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
		},
	},
	"Wireless": {
		{
			title: "WiFi Interfaces",
			path:  "/interface/wireless",
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
