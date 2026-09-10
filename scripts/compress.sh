#!/bin/sh
rm -f mistatic_deploy.tar
tar -cvf mistatic_deploy.tar \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='backend/pb_data' \
    --exclude='frontend/build' \
    --exclude='frontend/.svelte-kit' \
    --exclude='backend/mistatic' \
    --exclude='./sites' \
    .
