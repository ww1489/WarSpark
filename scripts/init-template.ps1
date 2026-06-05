param(
    [Parameter(Mandatory = $true)]
    [string]$ModulePath,

    [Parameter(Mandatory = $true)]
    [string]$Slug,

    [Parameter(Mandatory = $true)]
    [string]$DisplayName,

    [string]$EnvPrefix
)

$ErrorActionPreference = "Stop"

if ($Slug -notmatch '^[a-z][a-z0-9-]*$') {
    throw "Slug must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens."
}

if (-not $EnvPrefix) {
    $EnvPrefix = $Slug.ToUpperInvariant().Replace("-", "_")
}
if ($EnvPrefix -notmatch '^[A-Z][A-Z0-9_]*$') {
    throw "EnvPrefix must start with an uppercase letter and contain only uppercase letters, numbers, and underscores."
}

# Keep generated MySQL sample identifiers SQL-friendly when a service slug uses hyphens.
# This is not an init parameter; runtime config still owns mysql.user and mysql.db.
$databaseName = $Slug.Replace("-", "_")

$currentModulePath = "github.com/ww1489/WarSpark"
$currentSlug = "warspark"
$currentDisplayName = "WarSpark"
$currentEnvPrefix = "WARSPARK"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$repoRootPath = $repoRoot.Path.TrimEnd("\", "/")
Set-Location $repoRoot

$oldCommandPath = Join-Path $repoRoot "cmd\$currentSlug"
$newCommandPath = Join-Path $repoRoot "cmd\$Slug"
if ((Test-Path $oldCommandPath) -and ($Slug -ne $currentSlug)) {
    if (Test-Path $newCommandPath) {
        throw "Target command path already exists: $newCommandPath"
    }
    Move-Item -LiteralPath $oldCommandPath -Destination $newCommandPath
}

$replacements = @(
    @($currentModulePath, $ModulePath),
    @($currentDisplayName, $DisplayName),
    @($currentEnvPrefix, $EnvPrefix),
    @($currentSlug, $Slug)
)

$textExtensions = @(".go", ".mod", ".md", ".yaml", ".yml", ".ps1", ".sh", ".gitignore")
$textNames = @("Makefile", "LICENSE")
$files = Get-ChildItem -Path $repoRoot -Recurse -File -Force | Where-Object {
    $relativePath = $_.FullName.Substring($repoRootPath.Length).TrimStart("\", "/").Replace("\", "/")
    $isExcluded = $relativePath.StartsWith(".git/") -or
        $relativePath -eq "go.sum" -or
        $relativePath -eq "docs/swagger.json" -or
        $relativePath -eq "docs/swagger.yaml"
    $isText = ($textExtensions -contains $_.Extension) -or ($textNames -contains $_.Name)
    -not $isExcluded -and $isText
}

foreach ($file in $files) {
    $fullPath = $file.FullName
    $relativePath = $fullPath.Substring($repoRootPath.Length).TrimStart("\", "/").Replace("\", "/")
    $text = [System.IO.File]::ReadAllText($fullPath)
    foreach ($entry in $replacements) {
        $text = $text.Replace($entry[0], $entry[1])
    }
    $isInitScript = $relativePath -eq "scripts/init-template.ps1" -or $relativePath -eq "scripts/init-template.sh"
    # Normalize only sample/config defaults. Deployments override these through YAML or env vars.
    if (($databaseName -ne $Slug) -and -not $isInitScript) {
        $text = $text.Replace("user: $Slug", "user: $databaseName")
        $text = $text.Replace("db: $Slug", "db: $databaseName")
        $text = $text.Replace("DB:              AppName", "DB:              `"$databaseName`"")
        $text = $text.Replace("DB:              `"$Slug`"", "DB:              `"$databaseName`"")
    }
    $encoding = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($fullPath, $text, $encoding)
}

Write-Host "Template initialized:"
Write-Host "  module: $ModulePath"
Write-Host "  command: $Slug"
Write-Host "  display: $DisplayName"
Write-Host "  env prefix: $EnvPrefix"
Write-Host "  sample mysql identifiers: $databaseName (override in config)"
Write-Host ""
Write-Host "Next steps:"
Write-Host "  go mod tidy"
Write-Host "  go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/$Slug/main.go -o docs"
Write-Host "  go test -buildvcs=false ./..."
