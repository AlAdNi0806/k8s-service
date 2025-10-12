# Replace with your actual token
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjAzNzQ3NzUsInVzZXJfaWQiOjJ9.Jbt1GOx5O5kunzPsp66ar0ibuXHrfaXPlBjkCis1Eyc"

curl -X GET http://localhost:8082/health

# Using the JWT token in Authorization header
curl -X POST http://localhost:8082/orders \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 123, "quantity": 2}'
