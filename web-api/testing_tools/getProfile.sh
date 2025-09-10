#!/usr/bin/env bash

curl -X GET localhost:8080/api/profile \
-H "Content-Type: application/json" \
-H "Authorization: Bearer TOKEN" \
-d '{"email": "bruceape@gmail.com"}'
