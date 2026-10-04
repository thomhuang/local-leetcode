<#
.SYNOPSIS
    Launch local-leetcode with the DeepSeek API key pulled from Bitwarden.

.DESCRIPTION
    Reads the DeepSeek key from a Bitwarden vault item (default "DEEPSEEK API"),
    unlocking the vault when needed, sets DEEPSEEK_API_KEY for this process only,
    then runs the app from the repository root. The key is never written to disk.

    The key is read from a custom field named DEEPSEEK_API_TOKEN, then the
    password field, then the username field, then the item's notes.

.PARAMETER ItemName
    Bitwarden item name to read the key from. Default: "DEEPSEEK API".

.PARAMETER AppArgs
    Arguments forwarded to the app, for example -import-set neetcode-150.

.EXAMPLE
    ./scripts/run-with-bitwarden.ps1

.EXAMPLE
    ./scripts/run-with-bitwarden.ps1 -ItemName "DEEPSEEK API"
#>
[CmdletBinding()]
param(
    [string]$ItemName = "DEEPSEEK API",

    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$AppArgs
)

$ErrorActionPreference = "Stop"

# Get-BwItem finds a Bitwarden item by exact name or id, then falls back to a
# fuzzy search so a slightly different name (for example "DeepSeek API Key")
# still resolves.
function Get-BwItem([string]$Name) {
    $json = bw get item $Name --raw 2>$null
    if ($LASTEXITCODE -eq 0 -and $json) {
        return ($json | ConvertFrom-Json)
    }

    $json = bw list items --search $Name --raw 2>$null
    if ($LASTEXITCODE -eq 0 -and $json) {
        $matches = @($json | ConvertFrom-Json)
        if ($matches.Count -gt 0) {
            return $matches[0]
        }
    }

    return $null
}

# Get-DeepSeekKey extracts the key from the first field that carries it and
# reports where it came from. Custom fields win so an item can label the token.
function Get-DeepSeekKey($Item) {
    if ($Item.fields) {
        $named = $Item.fields | Where-Object { $_.value -and $_.name -match '(?i)^deepseek_api_token$' } | Select-Object -First 1
        if ($named) {
            return @{ value = $named.value; source = "custom field '$($named.name)'" }
        }
        $fuzzy = $Item.fields | Where-Object { $_.value -and $_.name -match '(?i)deepseek|api|token|key' } | Select-Object -First 1
        if ($fuzzy) {
            return @{ value = $fuzzy.value; source = "custom field '$($fuzzy.name)'" }
        }
    }
    if ($Item.login -and $Item.login.password) {
        return @{ value = $Item.login.password; source = "password field" }
    }
    if ($Item.login -and $Item.login.username) {
        return @{ value = $Item.login.username; source = "username field" }
    }
    if ($Item.notes) {
        return @{ value = $Item.notes.Trim(); source = "notes" }
    }
    return $null
}

# The app resolves all of its paths relative to the repository root.
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    if (-not (Get-Command bw -ErrorAction SilentlyContinue)) {
        throw "The Bitwarden CLI ('bw') was not found on PATH. Install it, then run 'bw login' once."
    }

    $status = (bw status | ConvertFrom-Json).status
    if ($status -eq "unauthenticated") {
        throw "Bitwarden is not logged in. Run 'bw login' first."
    }

    if ($env:BW_SESSION) {
        if ($status -eq "locked") {
            throw "Bitwarden is locked and BW_SESSION is stale. Run 'bw unlock' and set BW_SESSION again."
        }
    }
    else {
        Write-Host "Unlocking Bitwarden..." -ForegroundColor Yellow
        $session = bw unlock --raw
        if (-not $session) {
            throw "Could not obtain a Bitwarden session. Run 'bw unlock', set `$env:BW_SESSION to its output, then retry."
        }
        $env:BW_SESSION = "$session".Trim()
    }

    Write-Host "Syncing Bitwarden..." -ForegroundColor DarkGray
    bw sync 2>$null | Out-Null

    $item = Get-BwItem $ItemName
    if (-not $item) {
        $all = bw list items --raw 2>$null
        if ($LASTEXITCODE -eq 0 -and $all) {
            $names = @($all | ConvertFrom-Json | ForEach-Object { $_.name }) | Sort-Object -Unique
            $likely = @($names | Where-Object { $_ -match '(?i)deepseek|deep.?seek|api|key|token' })
            if ($likely.Count -gt 0) {
                throw "No Bitwarden item matched '$ItemName'. Closest names: $($likely -join ', ')."
            }
        }
        throw "No Bitwarden item matched '$ItemName'. Check the name with 'bw list items'."
    }

    $found = Get-DeepSeekKey $item
    if (-not $found) {
        throw "Item '$($item.name)' has no usable key (looked in custom fields, password, username, and notes)."
    }

    $env:DEEPSEEK_API_KEY = "$($found.value)".Trim()
    Write-Host "Injected DEEPSEEK_API_KEY from '$($item.name)' ($($found.source))." -ForegroundColor Green

    if ($AppArgs) {
        go run ./server @AppArgs
    }
    else {
        go run ./server
    }
}
finally {
    Remove-Item Env:DEEPSEEK_API_KEY -ErrorAction SilentlyContinue
    Pop-Location
}
