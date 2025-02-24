package template

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	model "github.com/nettica-com/nettica-admin/model"
)

var (
	emailTpl = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
    <!--[if gte mso 9]>
    <xml>
        <o:OfficeDocumentSettings>
            <o:AllowPNG/>
            <o:PixelsPerInch>96</o:PixelsPerInch>
        </o:OfficeDocumentSettings>
    </xml>
    <![endif]-->
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <meta name="format-detection" content="date=no" />
    <meta name="format-detection" content="address=no" />
    <meta name="format-detection" content="telephone=no" />
    <meta name="x-apple-disable-message-reformatting" />
    <!--[if !mso]><!-->
    <link href="https://fonts.googleapis.com/css?family=Muli:400,400i,700,700i" rel="stylesheet" />
    <!--<![endif]-->
    <title>Email Template</title>
    <!--[if gte mso 9]>
    <style type="text/css" media="all">
        sup { font-size: 100% !important; }
    </style>
    <![endif]-->
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">

    <style type="text/css" media="screen">
        /* Linked Styles */
        body { padding:0 !important; margin:0 !important; display:block !important; min-width:100% !important; width:100% !important; background:#001736; -webkit-text-size-adjust:none }
        a { color:#66c7ff; text-decoration:none }
        p { padding:0 !important; margin:0 !important }
        img { -ms-interpolation-mode: bicubic; /* Allow smoother rendering of resized image in Internet Explorer */ }
        .mcnPreviewText { display: none !important; }


        /* Mobile styles */
        @media only screen and (max-device-width: 480px), only screen and (max-width: 480px) {
            .mobile-shell { width: 100% !important; min-width: 100% !important; }
            .bg { background-size: 100% auto !important; -webkit-background-size: 100% auto !important; }

            .text-header,
            .m-center { text-align: center !important; }

            .center { margin: 0 auto !important; }
            .container { padding: 20px 10px !important }

            .td { width: 100% !important; min-width: 100% !important; }

            .m-br-15 { height: 15px !important; }
            .p30-15 { padding: 30px 15px !important; }

            .m-td,
            .m-hide { display: none !important; width: 0 !important; height: 0 !important; font-size: 0 !important; line-height: 0 !important; min-height: 0 !important; }

            .m-block { display: block !important; }

            .fluid-img img { width: 100% !important; max-width: 100% !important; height: auto !important; }

            .column,
            .column-top,
            .column-empty,
            .column-empty2,
            .column-dir-top { float: left !important; width: 100% !important; display: block !important; }

            .column-empty { padding-bottom: 10px !important; }
            .column-empty2 { padding-bottom: 30px !important; }

            .content-spacing { width: 15px !important; }
        }
    </style>
</head>
<body class="body" style="padding:0 !important; margin:0 !important; display:block !important; min-width:100% !important; width:100% !important; background:#001736; -webkit-text-size-adjust:none;">
<table width="100%" border="0" cellspacing="0" cellpadding="0" bgcolor="#001736">
    <tr>
        <td align="center" valign="top">
            <table width="650" border="0" cellspacing="0" cellpadding="0" class="mobile-shell">
                <tr>
                    <td class="td container" style="width:650px; min-width:650px; font-size:0pt; line-height:0pt; margin:0; font-weight:normal; padding:55px 0px;">

                        <!-- Article / Image On The Left - Copy On The Right -->
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td style="padding-bottom: 10px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td class="tbrr p30-15" style="padding: 60px 30px; border-radius:26px 26px 0px 0px;" bgcolor="#12325c">
                                                <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                                    <tr>
                                                        <th class="column-top" width="280" style="font-size:0pt; line-height:0pt; padding:0; margin:0; font-weight:normal; vertical-align:top;">
                                                            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                                                <tr>
                                                                    <td class="fluid-img" style="font-size:0pt; line-height:0pt; text-align:left;"><img src="cid:{{.QrcodePngName}}" width="280" height="210" border="0" alt="" /></td>
                                                                </tr>
                                                            </table>
                                                        </th>
                                                        <th class="column-empty2" width="30" style="font-size:0pt; line-height:0pt; padding:0; margin:0; font-weight:normal; vertical-align:top;"></th>
                                                        <th class="column-top" width="280" style="font-size:0pt; line-height:0pt; padding:0; margin:0; font-weight:normal; vertical-align:top;">
                                                            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                                                <tr>
                                                                    <td class="h4 pb20" style="color:#ffffff; font-family:'Muli', Arial,sans-serif; font-size:20px; line-height:28px; text-align:left; padding-bottom:20px;">Hi/td>
                                                                </tr>
                                                                <tr>
                                                                    <td class="text pb20" style="color:#ffffff; font-family:Arial,sans-serif; font-size:14px; line-height:26px; text-align:left; padding-bottom:20px;">Welcome to Nettica!  Click on the link to join the net.</td>
                                                                </tr>
                                                            </table>
                                                        </th>
                                                    </tr>
                                                </table>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <!-- END Article / Image On The Left - Copy On The Right -->

                        <!-- Two Columns / Articles -->
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td style="padding-bottom: 10px;">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0" bgcolor="#0e264b">
                                        <tr>
                                            <td>
                                                <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                                    <tr>
                                                        <td class="p30-15" style="padding: 50px 30px;">
                                                            <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                                                <tr>
                                                                    <td class="h3 pb20" style="color:#ffffff; font-family:'Muli', Arial,sans-serif; font-size:25px; line-height:32px; text-align:left; padding-bottom:20px;">About WireGuard</td>
                                                                </tr>
                                                                <tr>
                                                                    <td class="text pb20" style="color:#ffffff; font-family:Arial,sans-serif; font-size:14px; line-height:26px; text-align:left; padding-bottom:20px;">WireGuard is an extremely simple yet fast and modern VPN that utilizes state-of-the-art cryptography. It aims to be faster, simpler, leaner, and more useful than IPsec, while avoiding the massive headache. It intends to be considerably more performant than OpenVPN.</td>
                                                                </tr>
                                                                <!-- Button -->
                                                                <tr>
                                                                    <td align="left">
                                                                        <table border="0" cellspacing="0" cellpadding="0">
                                                                            <tr>
                                                                                <td class="blue-button text-button" style="background:#66c7ff; color:#c1cddc; font-family:'Muli', Arial,sans-serif; font-size:14px; line-height:18px; padding:12px 30px; text-align:center; border-radius:0px 22px 22px 22px; font-weight:bold;"><a href="https://www.wireguard.com/install" target="_blank" class="link-white" style="color:#ffffff; text-decoration:none;"><span class="link-white" style="color:#ffffff; text-decoration:none;">Download WireGuard VPN Client</span></a></td>
                                                                            </tr>
                                                                        </table>
                                                                    </td>
                                                                </tr>
                                                                <!-- END Button -->
                                                            </table>
                                                        </td>
                                                    </tr>
                                                </table>
                                            </td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <!-- END Two Columns / Articles -->

                        <!-- Footer -->
                        <table width="100%" border="0" cellspacing="0" cellpadding="0">
                            <tr>
                                <td class="p30-15 bbrr" style="padding: 50px 30px; border-radius:0px 0px 26px 26px;" bgcolor="#0e264b">
                                    <table width="100%" border="0" cellspacing="0" cellpadding="0">
                                        <tr>
                                            <td class="text-footer1 pb10" style="color:#c1cddc; font-family:'Muli', Arial,sans-serif; font-size:16px; line-height:20px; text-align:center; padding-bottom:10px;">Wg Gen Web - Simple Web based configuration generator for WireGuard</td>
                                        </tr>
                                        <tr>
                                            <td class="text-footer2" style="color:#8297b3; font-family:'Muli', Arial,sans-serif; font-size:12px; line-height:26px; text-align:center;"><a href="https://github.com/vx3r/wg-gen-web" target="_blank" class="link" style="color:#66c7ff; text-decoration:none;"><span class="link" style="color:#66c7ff; text-decoration:none;">More info on Github</span></a></td>
                                        </tr>
                                    </table>
                                </td>
                            </tr>
                        </table>
                        <!-- END Footer -->
                    </td>
                </tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>
`

	clientTpl = `[Interface]
Address = {{ StringsJoin .VPN.Current.Address ", " }}
PrivateKey = {{ .VPN.Current.PrivateKey }}
{{ if ne (len .Server.Dns) 0 -}}
DNS={{ StringsJoin .Server.Dns ", " }}
{{- end }}
{{ if ne .Server.Mtu 0 -}}
MTU = {{.Server.Mtu}}
{{- end}}
[Peer]
PublicKey = {{ .VPN.Current.PublicKey }}
PresharedKey = {{ .VPN.Current.PresharedKey }}
AllowedIPs = {{ StringsJoin .VPN.Current.AllowedIPs ", " }}
Endpoint = {{ .Server.Endpoint }}
PersistentKeepalive = {{.VPN.Current.PersistentKeepalive}}
`

	wireguardTemplate = `{{ if .VPN.Enable }}
# {{.VPN.Name }} / Updated: {{ .VPN.Updated }} / Created: {{ .VPN.Created }}
[Interface]
  {{- range .VPN.Current.Address }}
Address = {{ . }}
  {{- end }}
PrivateKey = {{ .VPN.Current.PrivateKey }}
{{ if ne .VPN.Current.ListenPort 0 -}}ListenPort = {{ .VPN.Current.ListenPort }}{{- end}}
{{ if .VPN.Current.Dns }}DNS={{ StringsJoin .VPN.Current.Dns ", " }}{{ end }}
{{ if ne .VPN.Current.Mtu 0 -}}MTU = {{.VPN.Current.Mtu}}{{- end}}
{{ if .VPN.Current.PreUp -}}PreUp = {{ .VPN.Current.PreUp }}{{- end}}
{{ if .VPN.Current.PostUp -}}PostUp = {{ .VPN.Current.PostUp }}{{- end}}
{{ if .VPN.Current.PreDown -}}PreDown = {{ .VPN.Current.PreDown }}{{- end}}
{{ if .VPN.Current.PostDown -}}PostDown = {{ .VPN.Current.PostDown }}{{- end}}
{{ range .VPNs }}
{{ if .Enable }}
# {{.Name}} / Updated: {{.Updated}} / Created: {{.Created}}
[Peer]
PublicKey = {{ .Current.PublicKey }}
PresharedKey = {{ .Current.PresharedKey }}
AllowedIPs = {{ StringsJoin .Current.AllowedIPs ", " }}
{{ if .Current.Endpoint -}}Endpoint = {{ .Current.Endpoint }} {{- end }}
{{ if .Current.PersistentKeepalive }}PersistentKeepalive = {{ .Current.PersistentKeepalive }}{{ end }}
{{ end }}
{{ end }}
{{ end }}`
)

// DumpWireguardConfig using go template
func DumpWireguardConfig(vpn *model.VPN, vpns []*model.VPN) ([]byte, error) {
	t, err := template.New("wireguard").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(wireguardTemplate)
	if err != nil {
		return nil, err
	}

	return dump(t, struct {
		VPN  *model.VPN
		VPNs []*model.VPN
	}{
		VPN:  vpn,
		VPNs: vpns,
	})
}

// DumpClientWg dump client wg config with go template
func DumpClientWg(vpn *model.VPN, server *model.Server) ([]byte, error) {
	t, err := template.New("client").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(clientTpl)
	if err != nil {
		return nil, err
	}

	return dump(t, struct {
		VPN    *model.VPN
		Server *model.Server
	}{
		VPN:    vpn,
		Server: server,
	})
}

// DumpEmail dump server wg config with go template
func DumpEmail(vpn *model.VPN, qrcodePngName string) ([]byte, error) {
	t, err := template.New("email").Parse(emailTpl)
	if err != nil {
		return nil, err
	}

	return dump(t, struct {
		VPN           *model.VPN
		QrcodePngName string
	}{
		VPN:           vpn,
		QrcodePngName: qrcodePngName,
	})
}

func BillingEmail(subscription model.Subscription) ([]byte, error) {
	file := "billing.html"

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	t, err := template.New("billing").Parse(string(bytes))
	if err != nil {
		return nil, err
	}

	return dump(t, struct {
		S model.Subscription
	}{
		S: subscription,
	})
}

// DumpEmail invites a user to join the network
func DumpUserEmail(accountId string) ([]byte, error) {
	file := "invite.html"
	server := os.Getenv("SERVER")

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	t, err := template.New("email").Parse(string(bytes))
	if err != nil {
		return nil, err
	}

	return dump(t, struct {
		AccountID string
		Server    string
	}{
		AccountID: accountId,
		Server:    server,
	})
}

func dump(tpl *template.Template, data interface{}) ([]byte, error) {
	var tplBuff bytes.Buffer

	err := tpl.Execute(&tplBuff, data)
	if err != nil {
		return nil, err
	}

	return tplBuff.Bytes(), nil
}
