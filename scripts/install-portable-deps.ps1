$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$tools = Join-Path $root ".tools"
$downloads = Join-Path $tools "_downloads"

New-Item -ItemType Directory -Force -Path $tools, $downloads | Out-Null

function Download-File {
  param(
    [string] $Url,
    [string] $OutFile
  )

  if (Test-Path $OutFile) {
    return
  }

  Write-Host "Downloading $Url"
  Invoke-WebRequest -Uri $Url -OutFile $OutFile
}

function Reset-Dir {
  param([string] $Path)
  if (Test-Path $Path) {
    Remove-Item -LiteralPath $Path -Recurse -Force
  }
  New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

$goJson = Invoke-RestMethod -Uri "https://go.dev/dl/?mode=json"
$goRelease = $goJson | Where-Object { $_.stable -eq $true } | Select-Object -First 1
$goFile = $goRelease.files | Where-Object {
  $_.os -eq "windows" -and $_.arch -eq "amd64" -and $_.kind -eq "archive"
} | Select-Object -First 1
if (-not $goFile) {
  throw "Could not find Go windows amd64 archive."
}

$goZip = Join-Path $downloads $goFile.filename
Download-File -Url "https://go.dev/dl/$($goFile.filename)" -OutFile $goZip
Reset-Dir (Join-Path $tools "go")
$goExtract = Join-Path $downloads "go-extract"
Reset-Dir $goExtract
Expand-Archive -LiteralPath $goZip -DestinationPath $goExtract -Force
Move-Item -LiteralPath (Join-Path $goExtract "go\*") -Destination (Join-Path $tools "go") -Force

$w64Url = "https://github.com/skeeto/w64devkit/releases/download/v2.5.0/w64devkit-x64-2.5.0.exe"
$w64Exe = Join-Path $downloads "w64devkit-x64-2.5.0.exe"
Download-File -Url $w64Url -OutFile $w64Exe
if (Test-Path (Join-Path $tools "w64devkit")) {
  Remove-Item -LiteralPath (Join-Path $tools "w64devkit") -Recurse -Force
}
& $w64Exe "-o$tools" -y
if (-not (Test-Path (Join-Path $tools "w64devkit\bin\gcc.exe"))) {
  throw "w64devkit extraction failed."
}

$ffmpegZip = Join-Path $downloads "ffmpeg-release-essentials.zip"
Download-File -Url "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip" -OutFile $ffmpegZip
Reset-Dir (Join-Path $tools "ffmpeg")
$ffmpegExtract = Join-Path $downloads "ffmpeg-extract"
Reset-Dir $ffmpegExtract
Expand-Archive -LiteralPath $ffmpegZip -DestinationPath $ffmpegExtract -Force
$ffmpegRoot = Get-ChildItem -LiteralPath $ffmpegExtract -Directory | Select-Object -First 1
if (-not $ffmpegRoot) {
  throw "FFmpeg extraction failed."
}
Move-Item -LiteralPath (Join-Path $ffmpegRoot.FullName "*") -Destination (Join-Path $tools "ffmpeg") -Force

$env:PATH = "$(Join-Path $tools "go\bin");$(Join-Path $tools "w64devkit\bin");$(Join-Path $tools "ffmpeg\bin");$env:PATH"

go version
gcc --version
ffmpeg -version
