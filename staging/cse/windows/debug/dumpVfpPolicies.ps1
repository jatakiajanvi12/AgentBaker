param(
   [string]$switchName = $(throw "please specify a switch name"),
   [string]$outfile = "vfprules.txt"
  )

$GithubSDNRepository = 'Microsoft/SDN'
if ((Test-Path env:GITHUB_SDN_REPOSITORY) -and ($env:GITHUB_SDN_REPOSITORY -ne ''))
{
    $GithubSDNRepository = $env:GITHUB_SDN_REPOSITORY
}

$BaseDir = "c:\k\debug"
md $BaseDir -ErrorAction Ignore

$helper = "$BaseDir\helper.psm1"

ipmo $helper
ipmo $BaseDir\VFP.psm1

$ports = Get-VfpPorts -SwitchName $switchName

# Dump the port info
$ports | select 'Port name', 'Mac Address', 'PortId' | Out-File $outfile -Encoding ascii -Append

$vfpCtrlExe = "vfpctrl.exe"

foreach ($port in $ports) {
	$portGuid = $port.'Port name'
	echo "Policy for port : " $portGuid  | Out-File $outfile -Encoding ascii -Append
	& $vfpCtrlExe /list-space  /port $portGuid | Out-File $outfile -Encoding ascii -Append
	& $vfpCtrlExe /list-mapping  /port $portGuid | Out-File $outfile -Encoding ascii -Append
	& $vfpCtrlExe /list-rule  /port $portGuid | Out-File $outfile -Encoding ascii -Append
	& $vfpCtrlExe /port $portGuid /get-port-state | Out-File $outfile -Encoding ascii -Append
	& $vfpCtrlExe /port $portGuid /list-nat-range | Out-File $outfile -Encoding ascii -Append
}

& $vfpCtrlExe /switch $ports[0].'Switch Name'  /get-switch-forwarding-settings > vswitchForwarding.txt