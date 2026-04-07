package test

import "time"

const (
	// ── ArgoCD sync / health timing ──────────────────────────────────────────

	syncMaxAttempts       = 5
	syncRetryInitialDelay = 10 * time.Second
	syncRetryDelay        = 30 * time.Second
	syncPollRetries       = 60
	syncPollInterval      = 10 * time.Second
	healthRetries         = 30
	healthInterval        = 10 * time.Second

	// ── Crossplane package kinds ──────────────────────────────────────────────

	FunctionKind      = "function.pkg.crossplane.io"
	ConfigurationKind = "configuration.pkg.crossplane.io"

	// ── Configuration names ───────────────────────────────────────────────────

	CronjobConfigurationName    = "platform-apis-cronjob"
	KafkaConfigurationName      = "platform-apis-kafka"
	PostgresqlConfigurationName = "platform-apis-postgresql"
	RepositoryConfigurationName = "platform-apis-repository"
	S3BucketConfigurationName   = "platform-apis-s3bucket"
	ValkeyConfigurationName     = "platform-apis-valkey"
	WebaccessConfigurationName  = "platform-apis-webaccess"
	WebappConfigurationName     = "platform-apis-webapp"
	ZoneConfigurationName       = "platform-apis-zone"

	// ── Function names ────────────────────────────────────────────────────────

	ArtifactFunctionName   = "platform-apis-artifact-fn"
	DatabaseFunctionName   = "platform-apis-database-fn"
	NetwokringFunctionName = "platform-apis-networking-fn"
	StorageFunctionName    = "platform-apis-storage-fn"
	TenancyFunctionName    = "platform-apis-tenancy-fn"
	WorkloadFunctionName   = "platform-apis-workload-fn"
	QueueFunctionName      = "platform-apis-queue-fn"

	// ── Zone Kyverno ─────────────────────────────────────────────────────

	ContributorKeyIDEnv  = "CONTRIBUTOR_AWS_ACCESS_KEY_ID"
	ContributorSecretEnv = "CONTRIBUTOR_AWS_SECRET_ACCESS_KEY"
	MaintainerKeyIDEnv   = "MAINTAINER_AWS_ACCESS_KEY_ID"
	MaintainerSecretEnv  = "MAINTAINER_AWS_SECRET_ACCESS_KEY"

	KyvernoTestNSName = "kyverno-test"

	// ── Zone ─────────────────────────────────────────────────────────────────

	ZoneApplicationName  = "app-of-zones"
	ZoneKind             = "zone.tenancy.entigo.com"
	NodeGroupKind        = "nodegroup.eks.aws.upbound.io"
	ZoneAName            = "a"
	ZoneBName            = "b"
	AAppsNamespace       = "a-apps"
	BAppsNamespace       = "b-apps"
	AAppsApplicationName = "a-apps"
	BAppsApplicationName = "b-apps"

	// ── CronJob ───────────────────────────────────────────────────────────────

	CronjobNamespaceName   = "test-cronjob"
	CronjobApplicationName = "test-cronjob"
	CronjobName            = "test-cronjob"
	CronjobKind            = "cronjobs.workload.entigo.com"
	BatchCronjobKind       = "cronjob"
	CronjobInitialSchedule = "0 * * * *"
	CronjobUpdatedSchedule = "30 * * * *"

	// ── Kafka ─────────────────────────────────────────────────────────────────

	KafkaNamespaceName   = "test-kafka"
	KafkaApplicationName = "test-kafka"

	KafkaClusterName        = "test-kafka-cluster"
	KafkaMSKObserverName    = KafkaClusterName + "-observed"
	KafkaMSKKind            = "msks.queue.entigo.com"
	KafkaClusterProvCfgKind = "clusterproviderconfig.kafka.m.crossplane.io"

	KafkaTopicName              = "test-topic-a"
	KafkaTopicKind              = "topics.queue.entigo.com"
	KafkaProvTopicKind          = "topic.topic.kafka.m.crossplane.io"
	KafkaTopicPartitions        = "6"
	KafkaTopicUpdatedPartitions = "9"
	KafkaTopicReplicationFactor = "3"

	KafkaUserName      = "test-user-a"
	KafkaUserKind      = "kafkausers.queue.entigo.com"
	KafkaACLKind       = "accesscontrollist.acl.kafka.m.crossplane.io"
	KafkaAWSSecKind    = "secret.secretsmanager.aws.m.upbound.io"
	KafkaAWSSecVerKind = "secretversion.secretsmanager.aws.m.upbound.io"
	KafkaAWSSecPolKind = "secretpolicy.secretsmanager.aws.m.upbound.io"
	KafkaSCRAMKind     = "singlescramsecretassociation.kafka.aws.m.upbound.io"

	// ── PostgreSQL ────────────────────────────────────────────────────────────

	PostgresqlNamespaceName   = "test-postgresql"
	PostgresqlApplicationName = "test-postgresql"

	PostgresqlInstanceName = "postgresql-instance-test"
	PostgresqlInstanceKind = "pginstances.database.entigo.com"
	RdsInstanceKind        = "instance.rds.aws.m.upbound.io"
	SecurityGroupKind      = "securitygroup.ec2.aws.m.upbound.io"
	SecurityGroupRuleKind  = "securitygrouprule.ec2.aws.m.upbound.io"
	ExternalSecretKind     = "externalsecret.external-secrets.io"
	SqlProviderConfigKind  = "providerconfig.postgresql.sql.m.crossplane.io"

	PostgresqlAdminUserName     = "test-owner"
	PostgresqlUserKind          = "postgresqlusers.database.entigo.com"
	PostgresqlAdminUserSpecName = "test_owner"
	PostgresqlRegularUserName   = "test-user"
	SqlRoleKind                 = "role.postgresql.sql.m.crossplane.io"

	RegularUserExpectedGrantName  = "grant-" + PostgresqlRegularUserName + "-test-owner-" + PostgresqlInstanceName
	RegularUserExpectedUsageName  = "usage-" + RegularUserExpectedGrantName
	RegularUserExpectedSecretName = PostgresqlInstanceName + "-" + PostgresqlRegularUserName

	AdminUserInstanceProtectionName   = PostgresqlAdminUserName + "-instance-protection"
	RegularUserInstanceProtectionName = PostgresqlRegularUserName + "-instance-protection"

	PostgresqlDatabaseKind = "postgresqldatabases.database.entigo.com"
	SqlDatabaseKind        = "database.postgresql.sql.m.crossplane.io"
	SqlExtensionKind       = "extension.postgresql.sql.m.crossplane.io"
	SqlGrantKind           = "grant.postgresql.sql.m.crossplane.io"
	UsageKind              = "usage.protection.crossplane.io"

	DatabaseOneName     = "database-one-test"
	DatabaseOneSpecName = "database_one_test"
	DatabaseTwoName     = "database-two-test"
	MinimalDatabaseName = "database-minimal-test"

	DatabaseGrantExpectedName    = DatabaseOneName + "-grant-owner-to-dbadmin"
	DatabaseTwoGrantExpectedName = DatabaseTwoName + "-grant-owner-to-dbadmin"

	DatabaseOneOwnerProtectionName     = DatabaseOneName + "-owner-protection"
	DatabaseTwoOwnerProtectionName     = DatabaseTwoName + "-owner-protection"
	MinimalDatabaseOwnerProtectionName = MinimalDatabaseName + "-owner-protection"

	DatabaseOneInstanceProtectionName     = DatabaseOneName + "-instance-protection"
	DatabaseTwoInstanceProtectionName     = DatabaseTwoName + "-instance-protection"
	MinimalDatabaseInstanceProtectionName = MinimalDatabaseName + "-instance-protection"

	// ── Repository ────────────────────────────────────────────────────────────

	RepositoryNamespaceName   = "test-repository"
	RepositoryApplicationName = "test-repository"

	RepositoryMinimalName       = "test-repo"
	RepositoryNamedName         = "test-repo-named"
	RepositoryNamedECRName      = "test-ecr-name"
	RepositoryNamedPath         = "test/path"
	RepositoryNamedExternalName = RepositoryNamedPath + "/" + RepositoryNamedECRName

	RepositoryKind    = "repositories.artifact.entigo.com"
	ECRRepositoryKind = "repository.ecr.aws.m.upbound.io"

	// ── S3 Bucket ─────────────────────────────────────────────────────────────

	S3BucketNamespaceName   = "test-s3bucket"
	S3BucketApplicationName = "test-s3bucket"

	S3MinimalName = "test-s3-minimal"

	S3BucketKind               = "s3buckets.storage.entigo.com"
	S3BucketAwsKind            = "bucket.s3.aws.m.upbound.io"
	S3IAMUserKind              = "user.iam.aws.m.upbound.io"
	S3IAMRoleKind              = "role.iam.aws.m.upbound.io"
	S3IAMPolicyKind            = "policy.iam.aws.m.upbound.io"
	S3SecretsManagerSecretKind = "secret.secretsmanager.aws.m.upbound.io"

	// ── Valkey ────────────────────────────────────────────────────────────────

	ValkeyNamespaceName   = "test-valkey"
	ValkeyApplicationName = "test-valkey"
	ValkeyCustomName      = "test-valkey-custom"

	ValkeyInstanceKind         = "valkeyinstances.database.entigo.com"
	ValkeyReplicationGroupKind = "replicationgroup.elasticache.aws.m.upbound.io"

	// ── WebApp ────────────────────────────────────────────────────────────────

	WebAppNamespaceName   = "test-webapp"
	WebAppApplicationName = "test-webapp"

	WebAppName           = "test-webapp"
	WebAppDeploymentName = "test-webapp"
	WebAppServiceName    = "test-webapp-service"
	WebAppSecretName     = "test-webapp-nginx-secret"

	WebAppKind = "webapps.workload.entigo.com"

	// ── WebAccess ─────────────────────────────────────────────────────────────

	WebAccessNamespaceName   = "test-webaccess"
	WebAccessApplicationName = "test-webaccess"

	WebAccessName = "test-webaccess"

	WebAccessKind             = "webaccesses.networking.entigo.com"
	WebAccessVirtualSvcKind   = "virtualservices.networking.istio.io"
	WebAccessServiceEntryKind = "serviceentries.networking.istio.io"
	WebAccessDestRuleKind     = "destinationrules.networking.istio.io"
)
