#!/usr/bin/env bash

curl -X POST localhost:8080/api/register \
-H "Content-Type: application/json" \
-d '{"email": "bruceape@gmail.com", "password": "opensaysme1"}'
