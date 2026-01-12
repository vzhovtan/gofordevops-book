package render

const JuniperTemplate = `## Juniper Configuration Generated from Infrastructure Model
## Device: {{.Hostname}}
## Model: {{.Model}}
## Generated: {{.Metadata.UpdatedAt.Format "2006-01-02 15:04:05"}}
##
system {
    host-name {{.Hostname}};
{{- if .Services.NTP.Enabled}}
    ntp {
{{- range .Services.NTP.Servers}}
        server {{.}};
{{- end}}
    }
{{- end}}
{{- if .Services.SNMP.Enabled}}
    snmp {
        community {{.Services.SNMP.Community}} {
            authorization read-only;
        }
        location "{{.Services.SNMP.Location}}";
        contact "{{.Services.SNMP.Contact}}";
    }
{{- end}}
{{- if .Services.Syslog.Enabled}}
    syslog {
{{- range .Services.Syslog.Servers}}
        host {{.Host}} {
            any {{.Severity}};
            port {{.Port}};
        }
{{- end}}
    }
{{- end}}
}
{{- if .VLANs}}
vlans {
{{- range .VLANs}}
    {{.Name}} {
        vlan-id {{.ID}};
        description "{{.Description}}";
    }
{{- end}}
}
{{- end}}
interfaces {
{{- range .Interfaces}}
    {{.Name}} {
        description "{{.Description}}";
{{- if .SwitchportMode}}
{{- if eq .SwitchportMode "access"}}
        unit 0 {
            family ethernet-switching {
                interface-mode access;
                vlan {
                    members {{.VLAN}};
                }
            }
        }
{{- else if eq .SwitchportMode "trunk"}}
        unit 0 {
            family ethernet-switching {
                interface-mode trunk;
                vlan {
                    members [ {{join .AllowedVLANs " "}} ];
                }
            }
        }
{{- end}}
{{- else}}
{{- if .IPAddress}}
        unit 0 {
            family inet {
                address {{.IPAddress}}/{{getMaskBits .SubnetMask}};
            }
        }
{{- end}}
{{- if .MTU}}
        mtu {{.MTU}};
{{- end}}
{{- if not .Enabled}}
        disable;
{{- end}}
{{- end}}
    }
{{- end}}
}
{{- if .Routing}}
{{- range .Routing.Protocols}}
{{- if eq .Protocol "ospf"}}
protocols {
    ospf {
        area {{range .Areas}}{{.AreaID}}{{end}} {
{{- range .Areas}}
{{- range $.Interfaces}}
            interface {{ .Name }} {
                interface-type p2p;
            }
{{- end}}
{{- end}}
        }
    }
}
{{- else if eq .Protocol "bgp"}}
protocols {
    bgp {
        group external {
            type external;
{{- range .Neighbors}}
            neighbor {{.IP}} {
                description "{{.Description}}";
                peer-as {{.RemoteAS}};
            }
{{- end}}
        }
    }
}
{{- end}}
{{- end}}
{{- if .Routing.StaticRoutes}}
routing-options {
    static {
{{- range .Routing.StaticRoutes}}
        route {{.Destination}} next-hop {{.NextHop}};
{{- end}}
    }
}
{{- end}}
{{- end}}`
