#!/bin/bash

# Load env vars from .env
source "$(git rev-parse --show-toplevel)/.env"

COMMIT=$(git log -1 --oneline)
curl -s --url 'smtps://smtp.gmail.com:465' \
  --user "$EMAIL_USER:$EMAIL_PASS" \
  --mail-from "$EMAIL_USER" \
  --mail-rcpt "$EMAIL_USER" \
  --upload-file <(echo -e "Subject: Push done\n\n$COMMIT")