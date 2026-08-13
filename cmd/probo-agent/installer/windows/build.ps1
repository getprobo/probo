# Copyright (c) 2026 Probo Inc <hello@probo.com>.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

#Requires -Version 7.0

<#
.SYNOPSIS
  Build the Probo Agent Windows MSI with WiX.

.DESCRIPTION
  Packages a pre-built probo-agent.exe into
  probo-agent_<version>_windows_<arch>.msi. Prefer passing an
  Authenticode-signed binary so the nested exe remains signed.

  Requires the WiX CLI (`wix`) on PATH (dotnet tool install --global wix).

.PARAMETER Binary
  Path to probo-agent.exe.

.PARAMETER Version
  Product version (X.Y.Z). Defaults to cmd/probo-agent/VERSION.

.PARAMETER Arch
  Target architecture: amd64, x86_64, x64, or arm64.

.PARAMETER Output
  Output .msi path. Defaults next to the script under dist/.
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Binary,

    [string]$Version = "",

    [Parameter(Mandatory = $true)]
    [ValidateSet("amd64", "x86_64", "x64", "arm64")]
    [string]$Arch,

    [string]$Output = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ScriptDir = $PSScriptRoot
$RepoRoot = (Resolve-Path (Join-Path $ScriptDir "..\..\..\..")).Path
$PackageWxs = Join-Path $ScriptDir "Package.wxs"

if (-not (Test-Path -LiteralPath $Binary)) {
    throw "error: --binary / -Binary path not found: $Binary"
}

$Binary = (Resolve-Path -LiteralPath $Binary).Path

if ([string]::IsNullOrWhiteSpace($Version)) {
    $VersionFile = Join-Path $RepoRoot "cmd\probo-agent\VERSION"
    if (-not (Test-Path -LiteralPath $VersionFile)) {
        throw "error: VERSION file missing at $VersionFile; pass -Version"
    }
    $Version = (Get-Content -LiteralPath $VersionFile -Raw).Trim()
}

if ($Version -notmatch '^\d+\.\d+\.\d+') {
    throw "error: version must look like X.Y.Z (got '$Version')"
}

switch ($Arch) {
    { $_ -in @("amd64", "x86_64", "x64") } {
        $WixArch = "x64"
        $ArchLabel = "x86_64"
    }
    "arm64" {
        $WixArch = "arm64"
        $ArchLabel = "arm64"
    }
}

if ([string]::IsNullOrWhiteSpace($Output)) {
    $DistDir = Join-Path $RepoRoot "dist"
    New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
    $Output = Join-Path $DistDir "probo-agent_${Version}_windows_${ArchLabel}.msi"
} else {
    $outDir = Split-Path -Parent $Output
    if ($outDir) {
        New-Item -ItemType Directory -Force -Path $outDir | Out-Null
    }
}

$wix = Get-Command wix -ErrorAction SilentlyContinue
if (-not $wix) {
    throw "error: wix CLI not found on PATH (install with: dotnet tool install --global wix)"
}

function ConvertTo-LicenseRtf {
    param([string]$LicensePath)

    $text = Get-Content -LiteralPath $LicensePath -Raw
    $escaped = $text.Replace('\', '\\').Replace('{', '\{').Replace('}', '\}')
    $escaped = $escaped -replace "`r`n", "\par`r`n"
    $escaped = $escaped -replace "(?<!\r)`n", "\par`n"
    return "{\rtf1\ansi\ansicpg1252\deff0{\fonttbl{\f0\fswiss Helvetica;}}\fs18 $escaped\par}"
}

function Ensure-WixUiExtension {
    $listed = & wix extension list -g 2>$null | Out-String
    if ($listed -notmatch 'WixToolset\.UI\.wixext') {
        Write-Host "Adding WixToolset.UI.wixext"
        & wix extension add -g WixToolset.UI.wixext/6.0.2
        if ($LASTEXITCODE -ne 0) {
            throw "error: failed to add WixToolset.UI.wixext"
        }
    }
}

Ensure-WixUiExtension

$LicenseSrc = Join-Path $RepoRoot "LICENSE"
if (-not (Test-Path -LiteralPath $LicenseSrc)) {
    throw "error: LICENSE missing at $LicenseSrc"
}

$LicenseRtf = Join-Path ([System.IO.Path]::GetTempPath()) "probo-agent-license-$PID.rtf"
[System.IO.File]::WriteAllText($LicenseRtf, (ConvertTo-LicenseRtf -LicensePath $LicenseSrc))
$LicenseRtfArg = $LicenseRtf.Replace('\', '/')

Write-Host "Building MSI: binary=$Binary arch=$WixArch version=$Version output=$Output"

try {
    & wix build `
        -arch $WixArch `
        -ext WixToolset.UI.wixext `
        -d "Version=$Version" `
        -d "AgentExe=$Binary" `
        -d "LicenseRtf=$LicenseRtfArg" `
        -o $Output `
        $PackageWxs

    if ($LASTEXITCODE -ne 0) {
        throw "error: wix build failed with exit code $LASTEXITCODE"
    }
} finally {
    Remove-Item -LiteralPath $LicenseRtf -ErrorAction SilentlyContinue
}

if (-not (Test-Path -LiteralPath $Output)) {
    throw "error: expected MSI was not produced at $Output"
}

# Package.wxs sets MediaTemplate EmbedCab=yes; fail loudly if WiX still
# emits a sidecar cabinet (users would need both files to install).
$outputDir = Split-Path -Parent $Output
if (-not $outputDir) {
    $outputDir = (Get-Location).Path
}
$sidecarCabs = @(Get-ChildItem -LiteralPath $outputDir -Filter "*.cab" -File -ErrorAction SilentlyContinue)
if ($sidecarCabs.Count -gt 0) {
    $names = ($sidecarCabs | ForEach-Object { $_.Name }) -join ", "
    throw "error: unexpected external cabinet(s) next to MSI ($names); expected EmbedCab=yes"
}

Write-Host "Wrote $Output"
