#!/usr/bin/env bash

curl -X GET localhost:8080/api/profile \
-H "Content-Type: application/json" \
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZjc2ZmZjNzctZDMzMi00ZjZjLWE4OGYtZDIzMjZiODZlZjc0IiwiZW1haWwiOiJicnVjZWFwZUBnbWFpbC5jb20iLCJleHAiOjE3NTc0NTg1MTksImlhdCI6MTc1NzM3MjExOX0.59clp_NQTQQgoZyMoiBSo6E62EmpNbtgpitWZx1IVdk" \
-d '{"email": "bruceape@gmail.com", "password": "opensaysme1"}'
