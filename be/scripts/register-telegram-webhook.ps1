param(
    [string]$WebhookUrl = "https://api.ppnetbet.com/payments/telegram-stars/webhook",
    [string]$ConfigPath = (Join-Path $PSScriptRoot "..\config.json"),
    [switch]$KeepPendingUpdates
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $ConfigPath -PathType Leaf)) {
    throw "Config file not found: $ConfigPath"
}

$resolvedConfigPath = (Resolve-Path -LiteralPath $ConfigPath).Path
$config = Get-Content -LiteralPath $resolvedConfigPath -Raw | ConvertFrom-Json
$telegram = $config.payment.telegram

if ($null -eq $telegram) {
    throw "Missing payment.telegram in $resolvedConfigPath"
}

$botToken = [string]$telegram.bot_token
$webhookSecret = [string]$telegram.webhook_secret

if ([string]::IsNullOrWhiteSpace($botToken) -or $botToken -like "PUT_*" -or $botToken -like "*_PLACEHOLDER") {
    throw "Please configure payment.telegram.bot_token in $resolvedConfigPath first."
}

if ([string]::IsNullOrWhiteSpace($webhookSecret) -or $webhookSecret -like "PUT_*" -or $webhookSecret -like "*_PLACEHOLDER") {
    throw "Please configure payment.telegram.webhook_secret in $resolvedConfigPath first."
}

if (-not [Uri]::IsWellFormedUriString($WebhookUrl, [UriKind]::Absolute) -or -not $WebhookUrl.StartsWith("https://")) {
    throw "Webhook URL must be a valid public HTTPS URL."
}

$setWebhookUri = "https://api.telegram.org/bot$botToken/setWebhook"
$getWebhookInfoUri = "https://api.telegram.org/bot$botToken/getWebhookInfo"
$body = @{
    url                  = $WebhookUrl
    secret_token         = $webhookSecret
    allowed_updates      = @("message", "pre_checkout_query")
    drop_pending_updates = -not $KeepPendingUpdates.IsPresent
} | ConvertTo-Json -Depth 4

$maskedToken = if ($botToken.Length -gt 10) {
    $botToken.Substring(0, 6) + "..." + $botToken.Substring($botToken.Length - 4)
} else {
    "***"
}

Write-Host "Registering Telegram webhook..."
Write-Host "Bot token: $maskedToken"
Write-Host "Webhook URL: $WebhookUrl"

$setResult = Invoke-RestMethod `
    -Method Post `
    -Uri $setWebhookUri `
    -ContentType "application/json" `
    -Body $body

if (-not $setResult.ok) {
    throw "Telegram rejected setWebhook: $($setResult.description)"
}

Write-Host "Webhook registration succeeded: $($setResult.description)" -ForegroundColor Green

$webhookInfo = Invoke-RestMethod -Method Get -Uri $getWebhookInfoUri
if (-not $webhookInfo.ok) {
    throw "Unable to read webhook information: $($webhookInfo.description)"
}

$result = $webhookInfo.result
Write-Host ""
Write-Host "Current webhook information:"
Write-Host "URL: $($result.url)"
Write-Host "Pending updates: $($result.pending_update_count)"
Write-Host "Last error: $($result.last_error_message)"

if ($result.url -ne $WebhookUrl) {
    throw "Webhook verification failed. Telegram returned a different URL: $($result.url)"
}

if (-not [string]::IsNullOrWhiteSpace([string]$result.last_error_message)) {
    Write-Warning "Telegram reports a webhook delivery error. Confirm that the API domain, HTTPS certificate, reverse proxy, and backend service are available."
}

Write-Host "Telegram webhook is configured." -ForegroundColor Green
