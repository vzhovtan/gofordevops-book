package infra

const CiscoTemplate = `!
! Cisco Configuration Generated from Infrastructure Model
! Device: {{.Hostname}}
! Model: {{.Model}}
! Generated: {{.Metadata.UpdatedAt.Format "2006-01-02 15:04:05"}}
!
hostname {{.Hostname}}
!
{{- if .Services.NTP.Enabled}}
{{range .Services.NTP.Servers}}
ntp server {{.}}
{{- end}}
{{- end}}
!
{{- if .Services.SNMP.Enabled}}
snmp-server community {{.Services.SNMP.Community}} RO
snmp-server location {{.Services.SNMP.Location}}
snmp-server contact {{.Services.SNMP.Contact}}
{{- end}}
!
{{- if .Services.Syslog.Enabled}}
{{range .Services.Syslog.Servers}}
logging host {{.Host}}
logging trap {{.Severity}}
{{- end}}
{{- end}}
!
{{- range .Interfaces}}
interface {{.Name}}
 description {{.Description}}
{{- if .IPAddress}}
 ip address {{.IPAddress}} {{.SubnetMask}}
{{- end}}
{{- if .MTU}}
 mtu {{.MTU}}
{{- end}}
 speed {{.Speed}}
 duplex {{.Duplex}}
{{- if .Enabled}}
 no shutdown
{{- else}}
 shutdown
{{- end}}
!
{{- end}}
{{- if .Routing}}
{{- range .Routing.Protocols}}
{{- if eq .Protocol "ospf"}}
router ospf {{.ProcessID}}
 router-id {{.RouterID}}
{{- range .Areas}}
{{ $areaid := .AreaID }}
{{- range .Networks}}
 network {{ . }} area {{ $areaid }}
{{- end}}
{{- end}}
!
{{- else if eq .Protocol "bgp"}}
router bgp {{.ASNumber}}
 bgp router-id {{.RouterID}}
{{- range .Neighbors}}
 neighbor {{.IP}} remote-as {{.RemoteAS}}
 neighbor {{.IP}} description {{.Description}}
{{- end}}
!
{{- end}}
{{- end}}
{{- if .Routing.StaticRoutes}}
{{range .Routing.StaticRoutes}}
ip route {{.Destination}} {{.NextHop}} {{.AdministrativeDistance}}
{{- end}}
!
{{- end}}
{{- end}}
end`
