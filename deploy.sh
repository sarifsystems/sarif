#!/usr/bin/env bash
set -eu

rm -rf out || exit 0;
mkdir out;

cd out
go get -v github.com/gopherjs/gopherjs
gopherjs build -m github.com/sarifsystems/sarif/cmd/util/js
mv js.js sarifserver.js
mv js.js.map sarifserver.js.map

git init
git config user.name "Travis CI"
git config user.email "me+travis@cschomburg.com"

git add .
git commit -m "Deploy to GitHub Pages"

git push --force --quiet "https://${GH_TOKEN}@github.com/sarifsystems/sarif.git" master:gh-pages > /dev/null 2>&1
