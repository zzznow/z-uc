# 在有网络的机器上运行，下载 tzdata 和 ca-certificates 及其依赖到本地
# 前提：本机已安装 Docker

$pkgsDir = "$PSScriptRoot\x86_64"
New-Item -ItemType Directory -Force -Path $pkgsDir | Out-Null

docker run --rm -v "${pkgsDir}:/out" alpine:3.23 sh -c `
    "apk fetch --no-cache -R -o /out tzdata ca-certificates && apk index -o /out/APKINDEX.tar.gz /out/*.apk"

Write-Output "Done. Packages downloaded to: $pkgsDir"
