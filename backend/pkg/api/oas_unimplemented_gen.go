// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"

	ht "github.com/ogen-go/ogen/http"
)

// UnimplementedHandler is no-op Handler which returns http.ErrNotImplemented.
type UnimplementedHandler struct{}

var _ Handler = UnimplementedHandler{}

// DatasourceEmailCreate implements datasource-email-create operation.
//
// Create a new email datasource.
//
// POST /datasource/email
func (UnimplementedHandler) DatasourceEmailCreate(ctx context.Context, req *DatasourceEmail) (r *DatasourceEmail, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceEmailDelete implements datasource-email-delete operation.
//
// Delete an email datasource.
//
// DELETE /datasource/email/{uuid}
func (UnimplementedHandler) DatasourceEmailDelete(ctx context.Context, params DatasourceEmailDeleteParams) error {
	return ht.ErrNotImplemented
}

// DatasourceEmailGet implements datasource-email-get operation.
//
// Get email datasources.
//
// GET /datasource/email/{uuid}
func (UnimplementedHandler) DatasourceEmailGet(ctx context.Context, params DatasourceEmailGetParams) (r *DatasourceEmail, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceEmailList implements datasource-email-list operation.
//
// List email datasources.
//
// GET /datasource/email
func (UnimplementedHandler) DatasourceEmailList(ctx context.Context, params DatasourceEmailListParams) (r []DatasourceEmail, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceEmailRunPipeline implements datasource-email-run-pipeline operation.
//
// Run datasource email pipeline.
//
// POST /datasource/email/{uuid}/run/pipeline
func (UnimplementedHandler) DatasourceEmailRunPipeline(ctx context.Context, params DatasourceEmailRunPipelineParams) (r *DatasourceEmailRunPipelineOK, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceEmailUpdate implements datasource-email-update operation.
//
// Update an email datasource.
//
// PUT /datasource/email/{uuid}
func (UnimplementedHandler) DatasourceEmailUpdate(ctx context.Context, req *DatasourceEmail, params DatasourceEmailUpdateParams) (r *DatasourceEmail, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceLinkedinCreate implements datasource-linkedin-create operation.
//
// Create a new LinkedIn datasource.
//
// POST /datasource/linkedin
func (UnimplementedHandler) DatasourceLinkedinCreate(ctx context.Context, req *DatasourceLinkedin) (r *DatasourceLinkedin, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceLinkedinDelete implements datasource-linkedin-delete operation.
//
// Delete a LinkedIn datasource.
//
// DELETE /datasource/linkedin/{uuid}
func (UnimplementedHandler) DatasourceLinkedinDelete(ctx context.Context, params DatasourceLinkedinDeleteParams) error {
	return ht.ErrNotImplemented
}

// DatasourceLinkedinGet implements datasource-linkedin-get operation.
//
// Get a LinkedIn datasource.
//
// GET /datasource/linkedin/{uuid}
func (UnimplementedHandler) DatasourceLinkedinGet(ctx context.Context, params DatasourceLinkedinGetParams) (r *DatasourceLinkedin, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceLinkedinList implements datasource-linkedin-list operation.
//
// List all LinkedIn datasources.
//
// GET /datasource/linkedin
func (UnimplementedHandler) DatasourceLinkedinList(ctx context.Context, params DatasourceLinkedinListParams) (r []DatasourceLinkedin, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceLinkedinUpdate implements datasource-linkedin-update operation.
//
// Update a LinkedIn datasource.
//
// PUT /datasource/linkedin/{uuid}
func (UnimplementedHandler) DatasourceLinkedinUpdate(ctx context.Context, req *DatasourceLinkedin, params DatasourceLinkedinUpdateParams) (r *DatasourceLinkedin, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceList implements datasource-list operation.
//
// Retrieve a list of datasource objects.
//
// GET /datasource
func (UnimplementedHandler) DatasourceList(ctx context.Context, params DatasourceListParams) (r []Datasource, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceSetOAuth2Client implements datasource-set-oauth2-client operation.
//
// Set OAuth2 client to the datasource.
//
// PUT /datasource/{uuid}/oauth2/client
func (UnimplementedHandler) DatasourceSetOAuth2Client(ctx context.Context, req *DatasourceSetOAuth2ClientReq, params DatasourceSetOAuth2ClientParams) error {
	return ht.ErrNotImplemented
}

// DatasourceTelegramCreate implements datasource-telegram-create operation.
//
// Create a new Telegram datasource.
//
// POST /datasource/telegram
func (UnimplementedHandler) DatasourceTelegramCreate(ctx context.Context, req *DatasourceTelegram) (r *DatasourceTelegram, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceTelegramDelete implements datasource-telegram-delete operation.
//
// Delete a Telegram datasource.
//
// DELETE /datasource/telegram/{uuid}
func (UnimplementedHandler) DatasourceTelegramDelete(ctx context.Context, params DatasourceTelegramDeleteParams) error {
	return ht.ErrNotImplemented
}

// DatasourceTelegramGet implements datasource-telegram-get operation.
//
// Get a Telegram datasource.
//
// GET /datasource/telegram/{uuid}
func (UnimplementedHandler) DatasourceTelegramGet(ctx context.Context, params DatasourceTelegramGetParams) (r *DatasourceTelegram, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceTelegramList implements datasource-telegram-list operation.
//
// List all Telegram datasources.
//
// GET /datasource/telegram
func (UnimplementedHandler) DatasourceTelegramList(ctx context.Context, params DatasourceTelegramListParams) (r []DatasourceTelegram, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceTelegramUpdate implements datasource-telegram-update operation.
//
// Update a Telegram datasource.
//
// PUT /datasource/telegram/{uuid}
func (UnimplementedHandler) DatasourceTelegramUpdate(ctx context.Context, req *DatasourceTelegram, params DatasourceTelegramUpdateParams) (r *DatasourceTelegram, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceWhatsappCreate implements datasource-whatsapp-create operation.
//
// Create a new WhatsApp datasource.
//
// POST /datasource/whatsapp
func (UnimplementedHandler) DatasourceWhatsappCreate(ctx context.Context, req *DatasourceWhatsapp) (r *DatasourceWhatsapp, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceWhatsappDelete implements datasource-whatsapp-delete operation.
//
// Delete a WhatsApp datasource.
//
// DELETE /datasource/whatsapp/{uuid}
func (UnimplementedHandler) DatasourceWhatsappDelete(ctx context.Context, params DatasourceWhatsappDeleteParams) error {
	return ht.ErrNotImplemented
}

// DatasourceWhatsappGet implements datasource-whatsapp-get operation.
//
// Get a WhatsApp datasource.
//
// GET /datasource/whatsapp/{uuid}
func (UnimplementedHandler) DatasourceWhatsappGet(ctx context.Context, params DatasourceWhatsappGetParams) (r *DatasourceWhatsapp, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceWhatsappList implements datasource-whatsapp-list operation.
//
// List all WhatsApp datasources.
//
// GET /datasource/whatsapp
func (UnimplementedHandler) DatasourceWhatsappList(ctx context.Context, params DatasourceWhatsappListParams) (r []DatasourceWhatsapp, _ error) {
	return r, ht.ErrNotImplemented
}

// DatasourceWhatsappUpdate implements datasource-whatsapp-update operation.
//
// Update a WhatsApp datasource.
//
// PUT /datasource/whatsapp/{uuid}
func (UnimplementedHandler) DatasourceWhatsappUpdate(ctx context.Context, req *DatasourceWhatsapp, params DatasourceWhatsappUpdateParams) (r *DatasourceWhatsapp, _ error) {
	return r, ht.ErrNotImplemented
}

// FileCreate implements file-create operation.
//
// Upload a new file and create its record.
//
// POST /file
func (UnimplementedHandler) FileCreate(ctx context.Context, req *UploadFileRequest) (r *UploadFileResponse, _ error) {
	return r, ht.ErrNotImplemented
}

// FileDelete implements file-delete operation.
//
// Delete a stored file.
//
// DELETE /file/{uuid}
func (UnimplementedHandler) FileDelete(ctx context.Context, params FileDeleteParams) error {
	return ht.ErrNotImplemented
}

// FileGet implements file-get operation.
//
// Retrieve details of a stored file.
//
// GET /file/{uuid}
func (UnimplementedHandler) FileGet(ctx context.Context, params FileGetParams) (r *FileObject, _ error) {
	return r, ht.ErrNotImplemented
}

// FileList implements file-list operation.
//
// Retrieve a list of stored files.
//
// GET /file
func (UnimplementedHandler) FileList(ctx context.Context, params FileListParams) (r []FileObject, _ error) {
	return r, ht.ErrNotImplemented
}

// FileUpdate implements file-update operation.
//
// Update metadata of a stored file.
//
// PUT /file/{uuid}
func (UnimplementedHandler) FileUpdate(ctx context.Context, req *FileUpdateReq, params FileUpdateParams) (r *FileObject, _ error) {
	return r, ht.ErrNotImplemented
}

// GenerateDownloadLink implements generateDownloadLink operation.
//
// Generate a download link for a stored file.
//
// POST /storage/file-link
func (UnimplementedHandler) GenerateDownloadLink(ctx context.Context, req *GenerateDownloadLinkRequest) (r *GenerateDownloadLinkResponse, _ error) {
	return r, ht.ErrNotImplemented
}

// GeneratePresignedUploadUrl implements generatePresignedUploadUrl operation.
//
// Generate a pre-signed URL for file upload.
//
// POST /storage/upload-url
func (UnimplementedHandler) GeneratePresignedUploadUrl(ctx context.Context, req *UploadPresignedUrlRequest) (r *UploadPresignedUrlResponse, _ error) {
	return r, ht.ErrNotImplemented
}

// MessageEmailQuery implements messageEmailQuery operation.
//
// Execute a search query on email messages.
//
// POST /message/email/query
func (UnimplementedHandler) MessageEmailQuery(ctx context.Context, req *MessageQuery) (r *MessageEmailQueryOK, _ error) {
	return r, ht.ErrNotImplemented
}

// MessageLinkedinQuery implements messageLinkedinQuery operation.
//
// Execute a search query on LinkedIn messages.
//
// POST /message/linkedin/query
func (UnimplementedHandler) MessageLinkedinQuery(ctx context.Context, req *MessageQuery) (r *MessageLinkedinQueryOK, _ error) {
	return r, ht.ErrNotImplemented
}

// MessageTelegramQuery implements messageTelegramQuery operation.
//
// Execute a search query on Telegram messages.
//
// POST /message/telegram/query
func (UnimplementedHandler) MessageTelegramQuery(ctx context.Context, req *MessageQuery) (r *MessageTelegramQueryOK, _ error) {
	return r, ht.ErrNotImplemented
}

// MessageWhatsappQuery implements messageWhatsappQuery operation.
//
// Execute a search query on WhatsApp messages.
//
// POST /message/whatsapp/query
func (UnimplementedHandler) MessageWhatsappQuery(ctx context.Context, req *MessageQuery) (r *MessageWhatsappQueryOK, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientCallback implements oauth2-client-callback operation.
//
// Serve OAuth2 client callback.
//
// GET /oauth2/callback
func (UnimplementedHandler) OAuth2ClientCallback(ctx context.Context, params OAuth2ClientCallbackParams) (r *OAuth2ClientCallbackFound, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientCreate implements oauth2-client-create operation.
//
// Create OAuth2 client.
//
// POST /oauth2/client
func (UnimplementedHandler) OAuth2ClientCreate(ctx context.Context, req *OAuth2ClientCreateReq) (r *OAuth2Client, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientDelete implements oauth2-client-delete operation.
//
// Delete OAuth2 client.
//
// DELETE /oauth2/client/{id}
func (UnimplementedHandler) OAuth2ClientDelete(ctx context.Context, params OAuth2ClientDeleteParams) error {
	return ht.ErrNotImplemented
}

// OAuth2ClientGet implements oauth2-client-get operation.
//
// Get OAuth2 client details.
//
// GET /oauth2/client/{id}
func (UnimplementedHandler) OAuth2ClientGet(ctx context.Context, params OAuth2ClientGetParams) (r *OAuth2Client, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientList implements oauth2-client-list operation.
//
// List OAuth2 clients.
//
// GET /oauth2/client
func (UnimplementedHandler) OAuth2ClientList(ctx context.Context, params OAuth2ClientListParams) (r *OAuth2ClientListOK, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientLogin implements oauth2-client-login operation.
//
// Start OAuth2 login flow.
//
// POST /oauth2/login
func (UnimplementedHandler) OAuth2ClientLogin(ctx context.Context, req *OAuth2ClientLoginReq) (r *OAuth2ClientLoginOK, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientTokenDelete implements oauth2-client-token-delete operation.
//
// Delete OAuth2 client token.
//
// DELETE /oauth2/client/{datasource_uuid}/token/{uuid}
func (UnimplementedHandler) OAuth2ClientTokenDelete(ctx context.Context, params OAuth2ClientTokenDeleteParams) error {
	return ht.ErrNotImplemented
}

// OAuth2ClientTokenList implements oauth2-client-token-list operation.
//
// List OAuth2 client tokens.
//
// GET /oauth2/client/{datasource_uuid}/token
func (UnimplementedHandler) OAuth2ClientTokenList(ctx context.Context, params OAuth2ClientTokenListParams) (r []OAuth2ClientToken, _ error) {
	return r, ht.ErrNotImplemented
}

// OAuth2ClientUpdate implements oauth2-client-update operation.
//
// Update OAuth2 client.
//
// PUT /oauth2/client/{id}
func (UnimplementedHandler) OAuth2ClientUpdate(ctx context.Context, req *OAuth2ClientUpdateReq, params OAuth2ClientUpdateParams) (r *OAuth2Client, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineCreate implements pipeline-create operation.
//
// Create Pipeline.
//
// POST /pipeline
func (UnimplementedHandler) PipelineCreate(ctx context.Context, req *PipelineCreateReq) (r *Pipeline, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineDelete implements pipeline-delete operation.
//
// Delete a pipeline.
//
// DELETE /pipeline/{uuid}
func (UnimplementedHandler) PipelineDelete(ctx context.Context, params PipelineDeleteParams) error {
	return ht.ErrNotImplemented
}

// PipelineEntryCreate implements pipeline-entry-create operation.
//
// Create a pipeline entry.
//
// POST /pipeline/{uuid}/entry
func (UnimplementedHandler) PipelineEntryCreate(ctx context.Context, req *PipelineEntryCreateReq, params PipelineEntryCreateParams) (r *PipelineEntry, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineEntryDelete implements pipeline-entry-delete operation.
//
// Delete pipeline entry.
//
// DELETE /pipeline/{uuid}/entry/{entry_uuid}
func (UnimplementedHandler) PipelineEntryDelete(ctx context.Context, params PipelineEntryDeleteParams) error {
	return ht.ErrNotImplemented
}

// PipelineEntryGet implements pipeline-entry-get operation.
//
// Get pipeline entry.
//
// GET /pipeline/{uuid}/entry/{entry_uuid}
func (UnimplementedHandler) PipelineEntryGet(ctx context.Context, params PipelineEntryGetParams) (r *PipelineEntry, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineEntryList implements pipeline-entry-list operation.
//
// Get pipeline entry.
//
// GET /pipeline/{uuid}/entry
func (UnimplementedHandler) PipelineEntryList(ctx context.Context, params PipelineEntryListParams) (r []PipelineEntry, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineEntryTypeList implements pipeline-entry-type-list operation.
//
// Get Pipeline Entry Types.
//
// GET /pipeline/entry/types
func (UnimplementedHandler) PipelineEntryTypeList(ctx context.Context) (r *PipelineEntryTypeListOK, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineEntryUpdate implements pipeline-entry-update operation.
//
// Update a pipeline entry.
//
// PUT /pipeline/{uuid}/entry/{entry_uuid}
func (UnimplementedHandler) PipelineEntryUpdate(ctx context.Context, req *PipelineEntryUpdateReq, params PipelineEntryUpdateParams) (r *PipelineEntry, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineGet implements pipeline-get operation.
//
// Get pipeline.
//
// GET /pipeline/{uuid}
func (UnimplementedHandler) PipelineGet(ctx context.Context, params PipelineGetParams) (r *Pipeline, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineList implements pipeline-list operation.
//
// Create Pipeline Object.
//
// GET /pipeline
func (UnimplementedHandler) PipelineList(ctx context.Context, params PipelineListParams) (r *PipelineListOK, _ error) {
	return r, ht.ErrNotImplemented
}

// PipelineUpdate implements pipeline-update operation.
//
// Update pipeline.
//
// PUT /pipeline/{uuid}
func (UnimplementedHandler) PipelineUpdate(ctx context.Context, req *PipelineUpdateReq, params PipelineUpdateParams) (r *Pipeline, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageHostfilesCreate implements storage-hostfiles-create operation.
//
// Create a new Host Files storage instance.
//
// POST /storage/hostfiles
func (UnimplementedHandler) StorageHostfilesCreate(ctx context.Context, req *StorageHostfiles) (r *StorageHostfiles, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageHostfilesDelete implements storage-hostfiles-delete operation.
//
// Delete a specific Host Files storage instance by UUID.
//
// DELETE /storage/hostfiles/{uuid}
func (UnimplementedHandler) StorageHostfilesDelete(ctx context.Context, params StorageHostfilesDeleteParams) error {
	return ht.ErrNotImplemented
}

// StorageHostfilesGet implements storage-hostfiles-get operation.
//
// Retrieve details of a specific Host Files storage instance by UUID.
//
// GET /storage/hostfiles/{uuid}
func (UnimplementedHandler) StorageHostfilesGet(ctx context.Context, params StorageHostfilesGetParams) (r *StorageHostfiles, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageHostfilesUpdate implements storage-hostfiles-update operation.
//
// Update details of a specific Host Files storage instance by UUID.
//
// PUT /storage/hostfiles/{uuid}
func (UnimplementedHandler) StorageHostfilesUpdate(ctx context.Context, req *StorageHostfiles, params StorageHostfilesUpdateParams) (r *StorageHostfiles, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageList implements storage-list operation.
//
// Retrieve a list of data storage objects.
//
// GET /storage
func (UnimplementedHandler) StorageList(ctx context.Context, params StorageListParams) (r []Storage, _ error) {
	return r, ht.ErrNotImplemented
}

// StoragePostgresCreate implements storage-postgres-create operation.
//
// Create a new PostgreSQL storage instance.
//
// POST /storage/postgres
func (UnimplementedHandler) StoragePostgresCreate(ctx context.Context, req *StoragePostgres) (r *StoragePostgres, _ error) {
	return r, ht.ErrNotImplemented
}

// StoragePostgresDelete implements storage-postgres-delete operation.
//
// Delete a specific PostgreSQL storage instance by UUID.
//
// DELETE /storage/postgres/{uuid}
func (UnimplementedHandler) StoragePostgresDelete(ctx context.Context, params StoragePostgresDeleteParams) error {
	return ht.ErrNotImplemented
}

// StoragePostgresGet implements storage-postgres-get operation.
//
// Retrieve details of a specific PostgreSQL storage instance by UUID.
//
// GET /storage/postgres/{uuid}
func (UnimplementedHandler) StoragePostgresGet(ctx context.Context, params StoragePostgresGetParams) (r *StoragePostgres, _ error) {
	return r, ht.ErrNotImplemented
}

// StoragePostgresUpdate implements storage-postgres-update operation.
//
// Update details of a specific PostgreSQL storage instance by UUID.
//
// PUT /storage/postgres/{uuid}
func (UnimplementedHandler) StoragePostgresUpdate(ctx context.Context, req *StoragePostgres, params StoragePostgresUpdateParams) (r *StoragePostgres, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageS3Create implements storage-s3-create operation.
//
// Create a new S3 storage instance.
//
// POST /storage/s3
func (UnimplementedHandler) StorageS3Create(ctx context.Context, req *StorageS3) (r *StorageS3, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageS3Delete implements storage-s3-delete operation.
//
// Delete a specific S3 storage instance by UUID.
//
// DELETE /storage/s3/{uuid}
func (UnimplementedHandler) StorageS3Delete(ctx context.Context, params StorageS3DeleteParams) error {
	return ht.ErrNotImplemented
}

// StorageS3Get implements storage-s3-get operation.
//
// Retrieve details of a specific S3 storage instance by UUID.
//
// GET /storage/s3/{uuid}
func (UnimplementedHandler) StorageS3Get(ctx context.Context, params StorageS3GetParams) (r *StorageS3, _ error) {
	return r, ht.ErrNotImplemented
}

// StorageS3Update implements storage-s3-update operation.
//
// Update details of a specific S3 storage instance by UUID.
//
// PUT /storage/s3/{uuid}
func (UnimplementedHandler) StorageS3Update(ctx context.Context, req *StorageS3, params StorageS3UpdateParams) (r *StorageS3, _ error) {
	return r, ht.ErrNotImplemented
}

// SyncpolicyCreate implements syncpolicy-create operation.
//
// Create a new sync policy.
//
// POST /syncpolicy
func (UnimplementedHandler) SyncpolicyCreate(ctx context.Context, req *SyncPolicy) (r *SyncPolicy, _ error) {
	return r, ht.ErrNotImplemented
}

// SyncpolicyDelete implements syncpolicy-delete operation.
//
// Delete a sync policy by uuid.
//
// DELETE /syncpolicy/{uuid}
func (UnimplementedHandler) SyncpolicyDelete(ctx context.Context, params SyncpolicyDeleteParams) error {
	return ht.ErrNotImplemented
}

// SyncpolicyGet implements syncpolicy-get operation.
//
// Retrieve a specific sync policy by uuid.
//
// GET /syncpolicy/{uuid}
func (UnimplementedHandler) SyncpolicyGet(ctx context.Context, params SyncpolicyGetParams) (r *SyncPolicy, _ error) {
	return r, ht.ErrNotImplemented
}

// SyncpolicyList implements syncpolicy-list operation.
//
// Retrieve a list of sync policies for the authenticated user.
//
// GET /syncpolicy
func (UnimplementedHandler) SyncpolicyList(ctx context.Context, params SyncpolicyListParams) (r *SyncpolicyListOK, _ error) {
	return r, ht.ErrNotImplemented
}

// SyncpolicyUpdate implements syncpolicy-update operation.
//
// Update a sync policy by uuid.
//
// PUT /syncpolicy/{uuid}
func (UnimplementedHandler) SyncpolicyUpdate(ctx context.Context, req *SyncPolicy, params SyncpolicyUpdateParams) (r *SyncPolicy, _ error) {
	return r, ht.ErrNotImplemented
}

// TgSessionCreate implements tg-session-create operation.
//
// Create a new Telegram session.
//
// POST /telegram
func (UnimplementedHandler) TgSessionCreate(ctx context.Context, req *TgSessionCreateReq) (r *Telegram, _ error) {
	return r, ht.ErrNotImplemented
}

// TgSessionList implements tg-session-list operation.
//
// List all Telegram sessions for the authenticated user.
//
// GET /telegram
func (UnimplementedHandler) TgSessionList(ctx context.Context) (r *TgSessionListOK, _ error) {
	return r, ht.ErrNotImplemented
}

// TgSessionVerify implements tg-session-verify operation.
//
// Complete the session creation process by verifying the code.
//
// PUT /telegram/{id}
func (UnimplementedHandler) TgSessionVerify(ctx context.Context, req *TgSessionVerifyReq, params TgSessionVerifyParams) (r *Telegram, _ error) {
	return r, ht.ErrNotImplemented
}

// UploadFile implements uploadFile operation.
//
// Upload a file.
//
// POST /storage/upload
func (UnimplementedHandler) UploadFile(ctx context.Context, req *UploadFileRequest) (r *UploadFileResponse, _ error) {
	return r, ht.ErrNotImplemented
}

// NewError creates *ErrorStatusCode from error returned by handler.
//
// Used for common default response.
func (UnimplementedHandler) NewError(ctx context.Context, err error) (r *ErrorStatusCode) {
	r = new(ErrorStatusCode)
	return r
}
