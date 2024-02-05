#!/bin/bash

npx tailwindcss -i public/css/input.css -o public/css/output.css --minify
templ generate

sleep 0.5

go build -o ./tmp/main .
