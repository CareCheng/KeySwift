package model

import "time"

// PluginRegistry 宿主管理的插件注册表。
type PluginRegistry struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	PluginID             string     `gorm:"size:120;uniqueIndex" json:"plugin_id"`
	InstallID            string     `gorm:"size:120;index" json:"install_id"`
	CurrentVersion       string     `gorm:"size:50" json:"current_version"`
	InstallRoot          string     `gorm:"size:500" json:"install_root"`
	SourceType           string     `gorm:"size:50" json:"source_type"`
	SourceURI            string     `gorm:"size:500" json:"source_uri"`
	Enabled              bool       `gorm:"default:false" json:"enabled"`
	Autostart            bool       `gorm:"default:false" json:"autostart"`
	DesiredState         string     `gorm:"size:50;default:approved-disabled" json:"desired_state"`
	ActualState          string     `gorm:"size:50;default:stopped" json:"actual_state"`
	LifecycleState       string     `gorm:"size:50;default:discovered" json:"lifecycle_state"`
	TrustLevel           string     `gorm:"size:50;default:local-approved" json:"trust_level"`
	SignatureStatus      string     `gorm:"size:50;default:unknown" json:"signature_status"`
	ApprovedManifestHash string     `gorm:"size:128" json:"approved_manifest_hash"`
	ApprovedPackageHash  string     `gorm:"size:128" json:"approved_package_hash"`
	ApprovedBinaryHash   string     `gorm:"size:128" json:"approved_binary_hash"`
	CurrentManifestHash  string     `gorm:"size:128" json:"current_manifest_hash"`
	LastVerifiedAt       *time.Time `json:"last_verified_at"`
	LastVerifyStatus     string     `gorm:"size:50" json:"last_verify_status"`
	TamperStatus         string     `gorm:"size:50" json:"tamper_status"`
	QuarantineReason     string     `gorm:"type:text" json:"quarantine_reason"`
	ConfigVersion        int        `gorm:"default:1" json:"config_version"`
	SelectedOS           string     `gorm:"size:50" json:"selected_os"`
	SelectedArch         string     `gorm:"size:50" json:"selected_arch"`
	HealthStatus         string     `gorm:"size:50;default:stopped" json:"health_status"`
	LastStartAt          *time.Time `json:"last_start_at"`
	LastReadyAt          *time.Time `json:"last_ready_at"`
	LastStopAt           *time.Time `json:"last_stop_at"`
	LastFaultAt          *time.Time `json:"last_fault_at"`
	ManifestJSON         string     `gorm:"type:text" json:"manifest_json"`
	ExtensionsJSON       string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PluginArtifact 记录插件包文件摘要与平台产物。
type PluginArtifact struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PluginID       string     `gorm:"size:120;index" json:"plugin_id"`
	ArtifactID     string     `gorm:"size:120;index" json:"artifact_id"`
	RelativePath   string     `gorm:"size:500" json:"relative_path"`
	ArtifactType   string     `gorm:"size:50" json:"artifact_type"`
	Platform       string     `gorm:"size:50" json:"platform"`
	Arch           string     `gorm:"size:50" json:"arch"`
	SizeBytes      int64      `json:"size_bytes"`
	HashAlgorithm  string     `gorm:"size:50" json:"hash_algorithm"`
	HashValue      string     `gorm:"size:128" json:"hash_value"`
	IsExecutable   bool       `gorm:"default:false" json:"is_executable"`
	IsRequired     bool       `gorm:"default:false" json:"is_required"`
	GroupName      string     `gorm:"size:100" json:"group_name"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PluginBinding 记录插件菜单、页面、路由、权限等宿主绑定关系。
type PluginBinding struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PluginID        string    `gorm:"size:120;index" json:"plugin_id"`
	BindingType     string    `gorm:"size:50;index" json:"binding_type"`
	BindingKey      string    `gorm:"size:180;index" json:"binding_key"`
	TargetScope     string    `gorm:"size:80" json:"target_scope"`
	MountArea       string    `gorm:"size:80" json:"mount_area"`
	RouteOrViewID   string    `gorm:"size:180" json:"route_or_view_id"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	OrderHint       int       `gorm:"default:0" json:"order_hint"`
	PermissionGuard string    `gorm:"size:200" json:"permission_guard"`
	ExtensionsJSON  string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PluginConfig 存储插件配置值与 schema 摘要。
type PluginConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PluginID        string    `gorm:"size:120;index" json:"plugin_id"`
	ConfigKey       string    `gorm:"size:180;index" json:"config_key"`
	ConfigVersion   int       `gorm:"default:1" json:"config_version"`
	SchemaJSON      string    `gorm:"type:text" json:"schema_json"`
	ValueJSON       string    `gorm:"type:text" json:"value_json"`
	EncryptedFields string    `gorm:"type:text" json:"encrypted_fields"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	UpdatedBy       string    `gorm:"size:120" json:"updated_by"`
	ExtensionsJSON  string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PluginEventLog 记录宿主与插件事件总线事件。
type PluginEventLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	EventID        string    `gorm:"size:120;uniqueIndex" json:"event_id"`
	EventType      string    `gorm:"size:120;index" json:"event_type"`
	EventVersion   string    `gorm:"size:50" json:"event_version"`
	OccurredAt     time.Time `gorm:"index" json:"occurred_at"`
	ProducerType   string    `gorm:"size:50" json:"producer_type"`
	ProducerID     string    `gorm:"size:120" json:"producer_id"`
	RequestID      string    `gorm:"size:120;index" json:"request_id"`
	SubjectID      string    `gorm:"size:120" json:"subject_id"`
	ResourceType   string    `gorm:"size:80;index" json:"resource_type"`
	ResourceID     string    `gorm:"size:120;index" json:"resource_id"`
	PayloadJSON    string    `gorm:"type:text" json:"payload_json"`
	ExtensionsJSON string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time `json:"created_at"`
}

// PluginJob 记录插件异步任务。
type PluginJob struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	JobID          string     `gorm:"size:120;uniqueIndex" json:"job_id"`
	OwnerPluginID  string     `gorm:"size:120;index" json:"owner_plugin_id"`
	JobType        string     `gorm:"size:120;index" json:"job_type"`
	ResourceType   string     `gorm:"size:80;index" json:"resource_type"`
	ResourceID     string     `gorm:"size:120;index" json:"resource_id"`
	Status         string     `gorm:"size:50;index;default:pending" json:"status"`
	Priority       int        `gorm:"default:0" json:"priority"`
	RunAt          time.Time  `gorm:"index" json:"run_at"`
	RetryCount     int        `gorm:"default:0" json:"retry_count"`
	MaxRetries     int        `gorm:"default:3" json:"max_retries"`
	LastError      string     `gorm:"type:text" json:"last_error"`
	PayloadJSON    string     `gorm:"type:text" json:"payload_json"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PluginMigration 记录插件结构操作声明和执行状态。
type PluginMigration struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PluginID       string     `gorm:"size:120;index" json:"plugin_id"`
	MigrationID    string     `gorm:"size:180;index" json:"migration_id"`
	Version        string     `gorm:"size:50" json:"version"`
	Direction      string     `gorm:"size:20" json:"direction"`
	Path           string     `gorm:"size:500" json:"path"`
	Checksum       string     `gorm:"size:128" json:"checksum"`
	Status         string     `gorm:"size:50;default:declared" json:"status"`
	ExecutedAt     *time.Time `json:"executed_at"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PluginDatabaseDeclaration 记录插件 manifest 的 database 顶层声明。
// 插件数据库接入规范见 Program/docs/Plugin_Development_Manual_CN/03-database-development.md。
type PluginDatabaseDeclaration struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PluginID       string    `gorm:"size:120;uniqueIndex" json:"plugin_id"`
	PluginVersion  string    `gorm:"size:50" json:"plugin_version"`
	Namespace      string    `gorm:"size:120;index" json:"namespace"`
	StorageMode    string    `gorm:"size:80;default:host-main-db" json:"storage_mode"`
	TableCount     int       `gorm:"default:0" json:"table_count"`
	Status         string    `gorm:"size:50;default:declared;index" json:"status"`
	ExtensionsJSON string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PluginDatabaseTable 记录插件 manifest 声明的数据库表。
// 插件数据库接入规范见 Program/docs/Plugin_Development_Manual_CN/03-database-development.md。
type PluginDatabaseTable struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	PluginID          string     `gorm:"size:120;index;uniqueIndex:idx_plugin_table_key" json:"plugin_id"`
	PluginVersion     string     `gorm:"size:50" json:"plugin_version"`
	Namespace         string     `gorm:"size:120;index" json:"namespace"`
	TableKey          string     `gorm:"size:120;uniqueIndex:idx_plugin_table_key" json:"table_key"`
	PhysicalTableName string     `gorm:"size:180;uniqueIndex" json:"physical_table_name"`
	TableKind         string     `gorm:"size:50;index" json:"table_kind"`
	SchemaVersion     string     `gorm:"size:50" json:"schema_version"`
	SchemaChecksum    string     `gorm:"size:128" json:"schema_checksum"`
	Status            string     `gorm:"size:50;default:declared;index" json:"status"`
	Sensitivity       string     `gorm:"size:50;default:internal" json:"sensitivity"`
	CreatePolicy      string     `gorm:"size:50;default:on_enable" json:"create_policy"`
	DropPolicy        string     `gorm:"size:50;default:manual_only" json:"drop_policy"`
	BackupPolicy      string     `gorm:"size:50;default:include" json:"backup_policy"`
	RetentionPolicy   string     `gorm:"size:100" json:"retention_policy"`
	Description       string     `gorm:"type:text" json:"description"`
	LastAppliedAt     *time.Time `json:"last_applied_at"`
	ExtensionsJSON    string     `gorm:"type:text" json:"extensions_json"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// PluginDatabaseColumn 记录插件表字段白名单。
type PluginDatabaseColumn struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	PluginID         string    `gorm:"size:120;index" json:"plugin_id"`
	TableID          uint      `gorm:"index;uniqueIndex:idx_plugin_column_name" json:"table_id"`
	ColumnKey        string    `gorm:"size:120" json:"column_key"`
	ColumnName       string    `gorm:"size:120;uniqueIndex:idx_plugin_column_name" json:"column_name"`
	DBType           string    `gorm:"size:80" json:"db_type"`
	LogicalType      string    `gorm:"size:80" json:"logical_type"`
	Nullable         bool      `gorm:"default:false" json:"nullable"`
	DefaultValueJSON string    `gorm:"type:text" json:"default_value_json"`
	PrimaryKey       bool      `gorm:"default:false" json:"primary_key"`
	AutoIncrement    bool      `gorm:"default:false" json:"auto_increment"`
	UniqueKey        bool      `gorm:"default:false" json:"unique_key"`
	Indexed          bool      `gorm:"default:false" json:"indexed"`
	Encrypted        bool      `gorm:"default:false" json:"encrypted"`
	Secret           bool      `gorm:"default:false" json:"secret"`
	ReferenceType    string    `gorm:"size:80" json:"reference_type"`
	ReferenceTarget  string    `gorm:"size:180" json:"reference_target"`
	Description      string    `gorm:"type:text" json:"description"`
	ExtensionsJSON   string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PluginDatabaseIndex 记录插件表索引声明。
type PluginDatabaseIndex struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PluginID       string    `gorm:"size:120;index" json:"plugin_id"`
	TableID        uint      `gorm:"index;uniqueIndex:idx_plugin_index_name" json:"table_id"`
	IndexKey       string    `gorm:"size:120" json:"index_key"`
	IndexName      string    `gorm:"size:180;uniqueIndex:idx_plugin_index_name" json:"index_name"`
	ColumnsJSON    string    `gorm:"type:text" json:"columns_json"`
	UniqueIndex    bool      `gorm:"default:false" json:"unique_index"`
	Status         string    `gorm:"size:50;default:declared" json:"status"`
	ExtensionsJSON string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PluginDatabaseRelation 记录插件表和宿主资源之间的逻辑引用。
type PluginDatabaseRelation struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	PluginID           string    `gorm:"size:120;index" json:"plugin_id"`
	TableID            uint      `gorm:"index;uniqueIndex:idx_plugin_relation_key" json:"table_id"`
	RelationKey        string    `gorm:"size:120;uniqueIndex:idx_plugin_relation_key" json:"relation_key"`
	LocalColumn        string    `gorm:"size:120" json:"local_column"`
	TargetResourceType string    `gorm:"size:120" json:"target_resource_type"`
	TargetKey          string    `gorm:"size:180" json:"target_key"`
	RelationType       string    `gorm:"size:50;default:many_to_one" json:"relation_type"`
	Required           bool      `gorm:"default:false" json:"required"`
	OnDeletePolicy     string    `gorm:"size:50;default:restrict" json:"on_delete_policy"`
	ExtensionsJSON     string    `gorm:"type:text" json:"extensions_json"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// PluginDatabaseOperation 记录插件数据库结构创建、校验和后续显式操作。
type PluginDatabaseOperation struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OperationID    string     `gorm:"size:120;uniqueIndex" json:"operation_id"`
	PluginID       string     `gorm:"size:120;index" json:"plugin_id"`
	PluginVersion  string     `gorm:"size:50" json:"plugin_version"`
	TableKey       string     `gorm:"size:120" json:"table_key"`
	OperationType  string     `gorm:"size:50" json:"operation_type"`
	Path           string     `gorm:"size:500" json:"path"`
	RequiresReview bool       `gorm:"default:false" json:"requires_review"`
	Status         string     `gorm:"size:50;default:pending;index" json:"status"`
	SchemaChecksum string     `gorm:"size:128" json:"schema_checksum"`
	ExecutedBy     string     `gorm:"size:120" json:"executed_by"`
	ExecutionMs    int        `gorm:"default:0" json:"execution_ms"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	ExtensionsJSON string     `gorm:"type:text" json:"extensions_json"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (PluginRegistry) TableName() string {
	return "plugin_registry"
}

func (PluginArtifact) TableName() string {
	return "plugin_artifacts"
}

func (PluginBinding) TableName() string {
	return "plugin_bindings"
}

func (PluginConfig) TableName() string {
	return "plugin_configs"
}

func (PluginEventLog) TableName() string {
	return "plugin_event_logs"
}

func (PluginJob) TableName() string {
	return "plugin_jobs"
}

func (PluginMigration) TableName() string {
	return "plugin_migrations"
}

func (PluginDatabaseDeclaration) TableName() string {
	return "plugin_database_declarations"
}

func (PluginDatabaseTable) TableName() string {
	return "plugin_database_tables"
}

func (PluginDatabaseColumn) TableName() string {
	return "plugin_database_columns"
}

func (PluginDatabaseIndex) TableName() string {
	return "plugin_database_indexes"
}

func (PluginDatabaseRelation) TableName() string {
	return "plugin_database_relations"
}

func (PluginDatabaseOperation) TableName() string {
	return "plugin_database_operations"
}
