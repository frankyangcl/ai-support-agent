param(
    [string]$DatabaseUrl = $env:DATABASE_URL,
    [string]$From = "006"
)
$ErrorActionPreference = "Stop"
if ([string]::IsNullOrWhiteSpace($DatabaseUrl)) { throw "DATABASE_URL is required" }
Get-ChildItem "$PSScriptRoot\..\migrations\*.sql" |
    Where-Object { $_.BaseName.Substring(0, 3) -ge $From } |
    Sort-Object Name | ForEach-Object {
    Write-Host "Applying $($_.Name)"
    & psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) { throw "Migration failed: $($_.Name)" }
}
