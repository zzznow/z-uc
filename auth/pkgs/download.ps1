$pkgsDir = "$PSScriptRoot\x86_64"
New-Item -ItemType Directory -Force -Path $pkgsDir | Out-Null

$base = "https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.23/main/x86_64"

# 1. 下载 APKINDEX（必须有，apk 才能找到包）
Write-Output "Downloading APKINDEX.tar.gz ..."
curl -L -o "$pkgsDir\APKINDEX.tar.gz" "$base/APKINDEX.tar.gz"

# 2. 下载需要的 .apk 包
$pkgs = @(
    "tzdata-2026b-r0",
    "ca-certificates-20260413-r0",
    "ca-certificates-bundle-20260413-r0"
)
foreach ($pkg in $pkgs) {
    $url = "$base/$pkg.apk"
    Write-Output "Downloading $pkg.apk ..."
    curl -L -o "$pkgsDir\$pkg.apk" $url
}

Write-Output "Done: $pkgsDir"
