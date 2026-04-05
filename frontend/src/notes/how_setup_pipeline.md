# How to Set Up a Pipeline

A **pipeline** connects a data source (where data comes from) to a storage target (where data goes). Follow these steps to get started.

## 1. Create a Storage Target

A storage target is where processed data will be written. ShadowAPI supports:

- **PostgreSQL** — write rows into database tables
- **S3** — write files to an S3-compatible bucket
- **Hostfiles** — write to local filesystem

Go to **Storages → Add Storage** and configure your target.

## 2. Add a Data Source

A data source defines where ShadowAPI pulls data from. Supported types:

- **Email (IMAP)** — connect via IMAP credentials
- **Email (OAuth2)** — connect via Google/Microsoft OAuth

Go to **Data Sources → Add Data Source** and provide connection details.

## 3. Create a Pipeline

Once you have at least one storage and one data source, create a pipeline to connect them:

- Select your **data source** (input)
- Select your **storage target** (output)
- Configure field mappings and processing options

Go to **Pipelines → Add Pipeline** to get started.

## 4. Run Your Pipeline

After creating a pipeline, trigger a run to start pulling and processing data:

- Go to **Pipelines** and click **Run** on your pipeline
- Monitor progress in the **Workers** section
- Once your first job completes, you're all set!
