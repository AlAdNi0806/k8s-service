TESTING ENDPOINTS:

# Register a new user
curl -X POST http://localhost:8081/register \
-H "Content-Type: application/json" \
-d '{"email": "testuser@example.com", "password": "securepassword"}'

# Login and get the JWT token
curl -X POST http://localhost:8081/login \
-H "Content-Type: application/json" \
-d '{"email": "testuser@example.com", "password": "securepassword"}'

# Test the protected endpoint using the token
TOKEN="eyJhbGciOiJIUzI1NiI..." # Replace with the actual token
curl -X GET http://localhost:8081/me \
-H "Authorization: Bearer ${TOKEN}"

# Test unauthorized access (e.g., without a token)
curl -X GET http://localhost:8081/me
# Expected: {"message":"Unauthorized"} (or similar) with status 401

# Check the health status
curl http://localhost:8081/health
