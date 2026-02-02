set -euo pipefail

appName="http-server"
version="v0.1.0"

list="
windows/amd64
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
"

# set version
sed -i -E "s|var version = \".+?\"|var version = \"${version}\"|" main.go

echo "${list}" | grep -v '^$' | grep -v '^#' | while read -r line; do
  os=$(echo "${line}" | cut -d '/' -f 1)
  arch=$(echo "${line}" | cut -d '/' -f 2)
  export GOOS=${os}
  export GOARCH=${arch}
  outputFile="${appName}-${version}-${GOOS}-${GOARCH}"
  echo "start package: ${outputFile}"
  rm -f http-server
  go build -o http-server
  chmod +x http-server
  tar -czf "${outputFile}.tgz" http-server
  rm -f http-server
done

sed -i -E "s|var version = \".+?\"|var version = \"<none>\"|" main.go

echo "package succeed"
