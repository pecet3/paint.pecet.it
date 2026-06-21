#!/bin/bash

npm run build &&
scp -P $SSH_PORT -i $SSH_KEY_PATH -r ./dist/*  $SSH_USER@$SSH_IP:$SSH_PBPUBLIC_PATH