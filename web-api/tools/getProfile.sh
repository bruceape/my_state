#!/usr/bin/env bash

curl -X GET localhost:8080/api/profile \
-H "Content-Type: application/json" \
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZjc2ZmZjNzctZDMzMi00ZjZjLWE4OGYtZDIzMjZiODZlZjc0IiwiZW1haWwiOiJicnVjZWFwZUBnbWFpbC5jb20iLCJleHAiOjE3NTc0NTcwODMsImlhdCI6MTc1NzM3MDY4M30.0R4TYLjhZ1hjV1_ZdCUyp75L704JWkB3_655dU_1HrE" \
-d '{"email": "bruceape@gmail.com", "password": "opensaysme1"}'
