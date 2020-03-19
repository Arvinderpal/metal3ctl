  #!/bin/bash

# Image url and checksum
export IMAGE_NAME='bionic-server-cloudimg-amd64.img'
export IMAGE_LOCATION='https://cloud-images.ubuntu.com/bionic/current'
export IMAGE_USERNAME='ubuntu'
export IMAGE_URL="http://172.22.0.1/images/{{ IMAGE_NAME }}"
export IMAGE_CHECKSUM="http://172.22.0.1/images/{{ IMAGE_NAME }}.md5sum"
