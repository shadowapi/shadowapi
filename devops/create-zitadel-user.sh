#!/bin/bash

# Script to create a new user in Zitadel using SQL
# Usage: ./create-zitadel-user.sh <email> <password> <firstname> <lastname>

EMAIL="${1:-test@example.com}"
PASSWORD="${2:-Test123!}"
FIRSTNAME="${3:-Test}"
LASTNAME="${4:-User}"

echo "Creating Zitadel user:"
echo "  Email: $EMAIL"
echo "  Name: $FIRSTNAME $LASTNAME"
echo ""

# Note: This is a simplified example. In production, Zitadel users should be created
# through the API or admin interface to properly handle password hashing and security.

cat << EOF
To create a user properly in Zitadel, you should:

1. Use the Zitadel Admin Console:
   - Login with: admin@shadowapi.localtest.me / Admin123!
   - Navigate to: http://localtest.me:8081
   - Go to Users section
   - Click "Add User"
   - Fill in the user details

2. Or use the Zitadel API:
   curl -X POST http://localtest.me:8081/management/v1/users/human/_import \\
     -H "Authorization: Bearer <admin-token>" \\
     -H "Content-Type: application/json" \\
     -d '{
       "userName": "$EMAIL",
       "profile": {
         "firstName": "$FIRSTNAME",
         "lastName": "$LASTNAME"
       },
       "email": {
         "email": "$EMAIL",
         "isEmailVerified": true
       },
       "password": "$PASSWORD"
     }'

3. Or use Zitadel CLI (if available in container):
   docker compose exec zitadel /app/zitadel admin setup user add \\
     --email "$EMAIL" \\
     --firstname "$FIRSTNAME" \\
     --lastname "$LASTNAME"

Note: Direct database manipulation is not recommended for creating users
as it bypasses password hashing, event sourcing, and security validations.
EOF