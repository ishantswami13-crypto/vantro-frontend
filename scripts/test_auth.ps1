param(
  [string]$Base = "http://localhost:8080"
)

function Format-Body {
  param([string]$Body)
  if ($null -eq $Body) { return "" }

  try {
    return ($Body | ConvertFrom-Json | ConvertTo-Json -Depth 20)
  } catch {
    return $Body
  }
}

function Invoke-Api {
  param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("GET", "POST", "PUT", "DELETE")]
    [string]$Method,

    [Parameter(Mandatory = $true)]
    [string]$Path,

    [object]$Body = $null
  )

  $url = "$Base$Path"
  $headers = @{ Accept = "application/json" }

  $jsonBody = $null
  if ($null -ne $Body) {
    $jsonBody = $Body | ConvertTo-Json -Depth 10
  }

  try {
    $resp = Invoke-WebRequest -Method $Method -Uri $url -Headers $headers -ContentType "application/json" -Body $jsonBody -ErrorAction Stop
    return [pscustomobject]@{
      Url        = $url
      StatusCode = [int]$resp.StatusCode
      Body       = $resp.Content
    }
  } catch {
    $statusCode = 0
    $content = $_.Exception.Message

    $response = $null
    if ($_.Exception.PSObject.Properties.Name -contains "Response") {
      $response = $_.Exception.Response
    }

    if ($null -ne $response) {
      try {
        if ($response -is [System.Net.Http.HttpResponseMessage]) {
          $statusCode = [int]$response.StatusCode
          $content = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        } else {
          if ($response.PSObject.Properties.Name -contains "StatusCode") {
            $statusCode = [int]$response.StatusCode
          }

          if ($response.PSObject.Properties.Name -contains "GetResponseStream") {
            $stream = $response.GetResponseStream()
            if ($null -ne $stream) {
              $reader = New-Object System.IO.StreamReader($stream)
              $content = $reader.ReadToEnd()
              $reader.Close()
            }
          }
        }
      } catch {
        # best-effort, keep defaults
      }
    }

    return [pscustomobject]@{
      Url        = $url
      StatusCode = $statusCode
      Body       = $content
    }
  }
}

function Print-Result {
  param(
    [string]$Label,
    $Result
  )

  Write-Host ""
  Write-Host "=== $Label ==="
  Write-Host "URL: $($Result.Url)"
  Write-Host "Status: $($Result.StatusCode)"
  Write-Host "Body:"
  Write-Host (Format-Body -Body $Result.Body)
}

Write-Host "Base: $Base"

Print-Result -Label "GET /" -Result (Invoke-Api -Method GET -Path "/")
Print-Result -Label "GET /health" -Result (Invoke-Api -Method GET -Path "/health")

$guid = [guid]::NewGuid().ToString("N")
$email = "test_$guid@example.com"
$password = "P@ssw0rd123!"

$signupBody = @{
  full_name = "Test User"
  email     = $email
  password  = $password
}

$loginBody = @{
  email    = $email
  password = $password
}

Print-Result -Label "POST /api/auth/signup" -Result (Invoke-Api -Method POST -Path "/api/auth/signup" -Body $signupBody)
Print-Result -Label "POST /api/auth/login" -Result (Invoke-Api -Method POST -Path "/api/auth/login" -Body $loginBody)

Print-Result -Label "GET /api/debug/users (requires DEBUG=true)" -Result (Invoke-Api -Method GET -Path "/api/debug/users")
