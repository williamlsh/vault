#!/usr/bin/env bash
curl -XPOST -d '{"password":"znm9832nmrfz4egwy43rn8"}' \
  -k \
  https://localhost:443/hash

curl -XPOST -d '{"password":"znm9832nmrfz4egwy43rn8","hash":"$2a$10$8e4JwCH9mCppJpTQ3Ax1PevFIt79her0oOg7AFy3eA4BNoeOMX1w."}' \
  -k \
  http://localhost:443/validate
