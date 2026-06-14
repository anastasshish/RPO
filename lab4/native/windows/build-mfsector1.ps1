# Build nfc-mfsector1.exe using MSYS2 MinGW (same toolchain as libnfc-build)
$ErrorActionPreference = 'Stop'
$Gcc = 'D:\msys2\mingw64\bin\gcc.exe'
if (-not (Test-Path $Gcc)) {
  $Gcc = (Get-Command gcc -ErrorAction SilentlyContinue).Source
}
if (-not $Gcc) {
  Write-Error 'gcc not found (install MSYS2 mingw64 or add to PATH)'
}

$LibnfcSrc = 'c:\src\libnfc'
$BuildDir = 'c:\src\libnfc-build'
$OutDir = $PSScriptRoot
$Src = Join-Path $OutDir 'src\nfc-mfsector1.c'
$MifareSrc = Join-Path $LibnfcSrc 'utils\mifare.c'
$exe = Join-Path $OutDir 'nfc-mfsector1.exe'

$inc = @(
  "-I$(Join-Path $LibnfcSrc 'include')",
  "-I$(Join-Path $BuildDir 'include')",
  "-I$(Join-Path $LibnfcSrc 'utils')"
)

$libnfc = Join-Path $BuildDir 'libnfc\libnfc.dll.a'
$utils = Join-Path $BuildDir 'utils\libnfcutils.a'

& $Gcc -O2 -o $exe $Src $MifareSrc @inc $libnfc $utils
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Built $exe"
