module gitlabe2.ext.net.nokia.com/atthegyi/netconf-client-go

go 1.18

require (
	github.com/go-logr/logr v1.2.0
	github.com/openshift-telco/go-netconf v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
)

require golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect

replace github.com/openshift-telco/go-netconf => github.com/openshift-telco/go-netconf-client v0.1.0
