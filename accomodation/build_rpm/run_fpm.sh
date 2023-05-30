set -e 

version="0.1.5"

fpm \
  -s dir -t rpm \
  -p handyman-$version-1-any.rpm \
  --name handyman \
  --license agpl3 \
  --version $version \
  --architecture all \
  --description "Go!" \
  handyman/main=/bin/handyman ../accomodation/handyman.service=/etc/systemd/system/handyman.service
