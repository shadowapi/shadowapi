## Obtain GMail API
This guide based on the [Google documentation](https://developers.google.com/gmail/api/guides).
To enable GMail API we need Google Workspace procject in the Google cloud:
 1. [Create Google Cloud](https://developers.google.com/workspace/guides/create-project) project.
 1. Enable API in the Product Library:
  1. Find [GMail API](https://console.cloud.google.com/workspace-api/products) and enable it.
1. Go to [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent):
  1. For **User type** select **Internal**
  1. Continue registration and save `credentials.json` in safe place.
