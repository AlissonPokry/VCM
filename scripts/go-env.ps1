param(
  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]] $CommandArgs
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$toolPaths = @(
  (Join-Path $root ".tools\go\bin"),
  (Join-Path $root ".tools\w64devkit\bin"),
  (Join-Path $root ".tools\ffmpeg\bin")
)

foreach ($path in $toolPaths) {
  if (Test-Path $path) {
    $env:PATH = "$path;$env:PATH"
  }
}

if ($CommandArgs.Count -eq 0) {
  throw "No command provided."
}

& $CommandArgs[0] @($CommandArgs | Select-Object -Skip 1)
exit $LASTEXITCODE
