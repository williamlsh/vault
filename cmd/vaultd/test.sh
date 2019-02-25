#!/usr/bin/env sh
curl -XPOST -d '{"password":"znm9832nmrfz4egwy43rn8"}' \
http://localhost:8080/hash

curl -XPOST -d '{"password":"znm9832nmrfz4egwy43rn8","hash":"$2a$10$8e4JwCH9mCppJpTQ3Ax1PevFIt79her0oOg7AFy3eA4BNoeOMX1w."}' \
http://localhost:8080/validate
