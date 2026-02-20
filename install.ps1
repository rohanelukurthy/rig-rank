$ErrorActionPreference = "Stop"

$repo = "rohanelukurthy/rig-rank"
$binName = "rigrank.exe"

Write-Host "Retrieving latest version of RigRank..."

# Determine Architecture
$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -eq "AMD64") {
    $archStr = "x86_64"
} elseif ($arch -eq "ARM64") {
    $archStr = "arm64"
} elseif ($arch -match "x86") {
    $archStr = "i386"
} else {
    Write-Host "Unsupported architecture: $arch" -ForegroundColor Red
    Exit
}

# Fetch latest release info
$releaseApiUrl = "https://api.github.com/repos/$repo/releases/latest"
Try {
    $release = Invoke-RestMethod -Uri $releaseApiUrl
    $tag = $release.tag_name
} Catch {
    Write-Host "Error fetching release information from GitHub API." -ForegroundColor Red
    Exit
}

Write-Host "Downloading $binName $tag (Windows-$archStr)..."

# Construct download URL (Note: Windows uses .zip in goreleaser overrides)
$zipName = "rigrank_Windows_$archStr.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$tag/$zipName"

$tempDir = Join-Path $env:TEMP "rigrank-install"
if (Test-Path $tempDir) { Remove-Item -Path $tempDir -Recurse -Force }
New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

$zipPath = Join-Path $tempDir "rigrank.zip"

Try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
} Catch {
    Write-Host "Error downloading release artifact. Does this architecture exist?" -ForegroundColor Red
    Remove-Item -Path $tempDir -Recurse -Force
    Exit
}

Write-Host "Extracting binary..."
Try {
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
} Catch {
    Write-Host "Error extracting ZIP file." -ForegroundColor Red
    Remove-Item -Path $tempDir -Recurse -Force
    Exit
}

$extractedExe = Join-Path $tempDir $binName

if (-not (Test-Path $extractedExe)) {
    Write-Host "Error: rigrank.exe not found in extracted archive." -ForegroundColor Red
    Remove-Item -Path $tempDir -Recurse -Force
    Exit
}

Write-Host "Installing $binName..."

# Install Location
$installDir = Join-Path $env:USERPROFILE ".rigrank\bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

$targetExe = Join-Path $installDir $binName
Move-Item -Path $extractedExe -Destination $targetExe -Force

# Add to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notmatch [regex]::Escape($installDir)) {
    Write-Host "Adding $installDir to user PATH..."
    $newPath = "$userPath;$installDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "PATH updated. You may need to restart your terminal to use 'rigrank' globally." -ForegroundColor Yellow
}

# Cleanup
Remove-Item -Path $tempDir -Recurse -Force

Write-Host "RigRank installation complete! The binary is located at $targetExe." -ForegroundColor Green
Write-Host "Run 'rigrank --help' to get started."
