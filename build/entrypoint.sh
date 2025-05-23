#!/bin/sh

mkdir -p data && cd data

for file in notify.js notify.py; do
    if [ ! -e "$file" ] || [ -L "$file" ]; then
        cp "../$file" .
    fi
done
printf "const notify = require('notify');\n\nnotify.sendNotify('标题', '内容');\n" > notify_sample.js
printf "import notify\n\nnotify.send('标题', '内容')\n" > notify_sample.py

if [ -f "npm.txt" ]; then
    npm install -g --no-cache $(cat npm.txt)
fi

if [ -f "pip.txt" ]; then
    pip install --no-cache-dir $(cat pip.txt)
fi

cd .. && ./xuanwu