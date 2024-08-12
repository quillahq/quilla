package constants

// DefaultDockerRegistry - default docker registry
const DefaultDockerRegistry = "https://index.docker.io"

// DefaultNamespace - default namespace to initialise configmaps based kv
const DefaultNamespace = "kube-system"

// WebhookEndpointEnv if set - enables webhook notifications
const WebhookEndpointEnv = "WEBHOOK_ENDPOINT"

// slack bot/token
const (
	EnvSlackToken            = "SLACK_TOKEN"
	EnvSlackBotName          = "SLACK_BOT_NAME"
	EnvSlackChannels         = "SLACK_CHANNELS"
	EnvSlackApprovalsChannel = "SLACK_APPROVALS_CHANNEL"

	EnvHipchatToken    = "HIPCHAT_TOKEN"
	EnvHipchatBotName  = "HIPCHAT_BOT_NAME"
	EnvHipchatChannels = "HIPCHAT_CHANNELS"

	EnvHipchatApprovalsChannel   = "HIPCHAT_APPROVALS_CHANNEL"
	EnvHipchatApprovalsUserName  = "HIPCHAT_APPROVALS_USER_NAME"
	EnvHipchatApprovalsBotName   = "HIPCHAT_APPROVALS_BOT_NAME"
	EnvHipchatApprovalsPasswort  = "HIPCHAT_APPROVALS_PASSWORT"
	EnvHipchatConnectionAttempts = "HIPCHAT_CONNECTION_ATTEMPTS"

	// Mattermost webhook endpoint, see https://docs.mattermost.com/developer/webhooks-incoming.html
	// for documentation on setting it up
	EnvMattermostEndpoint = "MATTERMOST_ENDPOINT"
	EnvMattermostName     = "MATTERMOST_USERNAME"

	// MS Teams webhook url, see https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using#setting-up-a-custom-incoming-webhook
	EnvTeamsWebhookUrl = "TEAMS_WEBHOOK_URL"

	// Discord webhook url, see https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
	EnvDiscordWebhookUrl = "DISCORD_WEBHOOK_URL"

	// Mail notification settings
	EnvMailTo         = "MAIL_TO"
	EnvMailFrom       = "MAIL_FROM"
	EnvMailSmtpServer = "MAIL_SMTP_SERVER"
	EnvMailSmtpPort   = "MAIL_SMTP_PORT"
	EnvMailSmtpUser   = "MAIL_SMTP_USER"
	EnvMailSmtpPass   = "MAIL_SMTP_PASS"
)

// EnvNotificationLevel - minimum level for notifications, defaults to info
const EnvNotificationLevel = "NOTIFICATION_LEVEL"

// Basic Auth - User / Password
const EnvBasicAuthUser = "BASIC_AUTH_USER"
const EnvBasicAuthPassword = "BASIC_AUTH_PASSWORD"
const EnvBasicAuthEnabled = "BASIC_AUTH_ENABLED"
const EnvAuthenticatedWebhooks = "AUTHENTICATED_WEBHOOKS"
const EnvTokenSecret = "TOKEN_SECRET"

// Azure
const EnvAzureTenantID = "AZURE_TENANT_ID"
const EnvAzureClientID = "AZURE_CLIENT_ID"
const EnvAzureClientSecret = "AZURE_CLIENT_SECRET"
const EnvAzureRedirectUri = "AZURE_REDIRECT_URI"
const EnvAzureAADEnabled = "AZURE_AAD_ENABLED"

// RBAC
const EnvRBACEnabled = "RBAC_ENABLED"
const EnvRBACPolicy = "RBAC_POLICY"

// quillaLogoURL - is a logo URL for bot icon
const QuillaLogoURL = "https://quilla.sh/img/logo.png"

// Env var to define a namespace that quilla will scan - avoid scan over all the cluster -
const EnvRestrictedNamespace = "RESTRICTED_NAMESPACE"
