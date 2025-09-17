# Zitadel Service User Setup

This document describes how to set up a service user for Zitadel Management API access.

## Prerequisites

1. Zitadel instance running
2. Admin access to Zitadel Console
3. Organization and Project already created

## Steps

### 1. Create Service User

1. Go to Zitadel Console: `http://auth.localtest.me/ui/console`
2. Login with admin credentials (`admin@example.com` / `Admin123!`)
3. Navigate to **Organization** → **Users**
4. Click **New** → **Service User**
5. Fill in the details:
   - **User Name**: `shadowapi-service`
   - **Name**: `ShadowAPI Service User`
   - **Description**: `Service user for ShadowAPI Management API access`

### 2. Generate Private Key

1. After creating the service user, click on it
2. Go to **Authentication** → **Private Key JWT**
3. Click **Add Key**
4. Choose key type (recommended: **RS256**)
5. Click **Add**
6. Download the JSON key file

### 3. Grant Permissions

1. Go to **Organization** → **Authorization**
2. Find your service user
3. Click **Grant Authorization**
4. Select roles:
   - `ORG_OWNER` (for organization-level operations)
   - `PROJECT_OWNER` (for project-level operations)

### 4. Configure ShadowAPI

1. Copy the downloaded key file to a secure location
2. Update environment variables:
   ```bash
   SA_AUTH_USER_MANAGER=zitadel
   SA_AUTH_ZITADEL_MANAGEMENT_URL=http://zitadel:8080
   SA_AUTH_ZITADEL_SERVICE_USER_ID=<user-id-from-key-file>
   SA_AUTH_ZITADEL_SERVICE_USER_KEY_PATH=/path/to/key.json
   ```

### 5. Test Authentication

The service user will automatically authenticate when the application starts.
Check the logs for any authentication errors.

## Key File Format

The downloaded key file should look like this:

```json
{
  "type": "serviceaccount",
  "keyId": "123456789",
  "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "userId": "123456789@shadowapi"
}
```

## Troubleshooting

### Common Issues

1. **"Invalid JWT"**: Check that the service user ID and key ID match
2. **"Insufficient permissions"**: Ensure the service user has the required roles
3. **"Key not found"**: Verify the key file path and format

### Debug Mode

Enable debug logging to see JWT authentication details:

```bash
SA_LOG_LEVEL=DEBUG
```

## Security Considerations

1. **Store key files securely** - never commit them to version control
2. **Rotate keys regularly** - generate new keys periodically
3. **Use least privilege** - only grant necessary permissions
4. **Monitor usage** - check logs for suspicious activity