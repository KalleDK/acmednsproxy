#!/bin/bash

echo h
export BASE_URL=${BASE_URL:-"https://kalledk.github.io/alpinerepo"}
export REPO_NAME=${REPO_NAME:-acmednsproxy}
export KEY_NAME=${KEY_NAME:-"alpine@k-moeller.dk-62068d1b.rsa.pub"}
export REPO_DEST=${REPO_DEST:-./packages}

cat << EOF > ${REPO_DEST}/${REPO_NAME}/index.md
# ACME DNS Proxy

\`\`\`bash
# Install key
wget -O "/etc/apk/keys/${KEY_NAME}" "${BASE_URL}/${REPO_NAME}/${KEY_NAME}"

# Install repo
echo "${BASE_URL}/${REPO_NAME}" >> /etc/apk/repositories
\`\`\` 
EOF
echo p