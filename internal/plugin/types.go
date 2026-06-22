// Package plugin 定义宿主插件体系的协议对象与注册接口。
package plugin

import (
	"context"
	"time"
)

const (
	ManifestVersion   = "1.0.0"
	BackendProtocol   = "1.0.0"
	FrontendProtocol  = "1.0.0"
	HandshakeProtocol = "1.0.0"
)

const (
	PluginKindFunctional  = "functional"
	PluginKindUITheme     = "ui-theme"
	PluginKindIntegration = "integration"
	PluginKindTooling     = "tooling"
)

const (
	RenderModeHostRendered = "host-rendered"
	RenderModeBundle       = "bundle"
	RenderModeIframe       = "iframe"
)

const (
	PluginStateDiscovered       = "discovered"
	PluginStateApprovedDisabled = "approved-disabled"
	PluginStateEnabled          = "enabled"
	PluginStateQuarantined      = "quarantined"
	PluginRuntimeStopped        = "stopped"
	PluginRuntimeStarting       = "starting"
	PluginRuntimeHandshaking    = "handshaking"
	PluginRuntimeRegistering    = "registering"
	PluginRuntimeReady          = "ready"
	PluginRuntimeDegraded       = "degraded"
	PluginRuntimeCrashed        = "crashed"
)

// ExtensionMap 承载未来未知能力的命名空间扩展字段。
type ExtensionMap map[string]any

// Manifest 是插件包 manifest.json 的宿主侧协议骨架。
type Manifest struct {
	ManifestVersion string                  `json:"manifestVersion"`
	ID              string                  `json:"id"`
	Version         string                  `json:"version"`
	PluginKind      string                  `json:"pluginKind"`
	Identity        Identity                `json:"identity"`
	Compatibility   Compatibility           `json:"compatibility"`
	Package         PackageInfo             `json:"package"`
	Integrity       Integrity               `json:"integrity"`
	Lifecycle       Lifecycle               `json:"lifecycle"`
	Dependencies    Dependencies            `json:"dependencies"`
	Permissions     []PermissionDeclaration `json:"permissions"`
	Capabilities    Capabilities            `json:"capabilities"`
	Backend         Backend                 `json:"backend"`
	Database        DatabaseDeclaration     `json:"database"`
	Frontend        Frontend                `json:"frontend"`
	UI              UIContribution          `json:"ui"`
	Observability   Observability           `json:"observability"`
	Operations      Operations              `json:"operations"`
	Metadata        map[string]any          `json:"metadata"`
	Extensions      ExtensionMap            `json:"extensions"`
}

// Identity 描述插件的人类可读身份。
type Identity struct {
	Name           string       `json:"name"`
	DisplayName    string       `json:"displayName"`
	Description    string       `json:"description"`
	Summary        string       `json:"summary"`
	Keywords       []string     `json:"keywords"`
	Categories     []string     `json:"categories"`
	Tags           []string     `json:"tags"`
	Author         string       `json:"author"`
	Organization   string       `json:"organization"`
	License        string       `json:"license"`
	Homepage       string       `json:"homepage"`
	Repository     string       `json:"repository"`
	Support        string       `json:"support"`
	ReleaseChannel string       `json:"releaseChannel"`
	CreatedAt      string       `json:"createdAt"`
	UpdatedAt      string       `json:"updatedAt"`
	Extensions     ExtensionMap `json:"extensions"`
}

// Compatibility 描述插件与宿主、协议和平台的兼容关系。
type Compatibility struct {
	HostApp                string       `json:"hostApp"`
	MinHostVersion         string       `json:"minHostVersion"`
	MaxHostVersion         string       `json:"maxHostVersion"`
	FrontendProtocol       string       `json:"frontendProtocol"`
	BackendProtocol        string       `json:"backendProtocol"`
	HandshakeProtocol      string       `json:"handshakeProtocol"`
	SupportedPlatforms     []string     `json:"supportedPlatforms"`
	SupportedArchitectures []string     `json:"supportedArchitectures"`
	RuntimeModes           []string     `json:"runtimeModes"`
	DatabaseModes          []string     `json:"databaseModes"`
	RequiredHostFeatures   []string     `json:"requiredHostFeatures"`
	DeprecatedFields       []string     `json:"deprecatedFields"`
	BreakingChangesSince   string       `json:"breakingChangesSince"`
	Extensions             ExtensionMap `json:"extensions"`
}

// PackageInfo 描述插件包入口、二进制和资源引用。
type PackageInfo struct {
	PackageFormat    string            `json:"packageFormat"`
	InstallMode      string            `json:"installMode"`
	DistributionType string            `json:"distributionType"`
	EntryStrategy    string            `json:"entryStrategy"`
	DefaultBinary    string            `json:"defaultBinary"`
	Binaries         []BinaryInfo      `json:"binaries"`
	ResourceRefs     map[string]string `json:"resourceRefs"`
	FrontendRef      string            `json:"frontendRef"`
	DocsRef          string            `json:"docsRef"`
	LocalesRef       string            `json:"localesRef"`
	AssetsRef        string            `json:"assetsRef"`
	OptionalGroups   []string          `json:"optionalGroups"`
	HotReloadPolicy  string            `json:"hotReloadPolicy"`
	PackageDigestRef string            `json:"packageDigestRef"`
	Extensions       ExtensionMap      `json:"extensions"`
}

// BinaryInfo 描述单个平台二进制。
type BinaryInfo struct {
	Platform   string       `json:"platform"`
	Arch       string       `json:"arch"`
	Path       string       `json:"path"`
	Entrypoint string       `json:"entrypoint"`
	Extensions ExtensionMap `json:"extensions"`
}

// Integrity 描述完整性校验和信任策略。
type Integrity struct {
	Enabled                    bool         `json:"enabled"`
	HashAlgorithm              string       `json:"hashAlgorithm"`
	ChecksumFile               string       `json:"checksumFile"`
	SignatureFile              string       `json:"signatureFile"`
	PackageDigest              string       `json:"packageDigest"`
	ManifestDigest             string       `json:"manifestDigest"`
	RequiredScopes             []string     `json:"requiredScopes"`
	RequireApprovedFingerprint bool         `json:"requireApprovedFingerprint"`
	VerifyOnInstall            bool         `json:"verifyOnInstall"`
	VerifyOnEnable             bool         `json:"verifyOnEnable"`
	VerifyBeforeStart          bool         `json:"verifyBeforeStart"`
	BackgroundAuditPolicy      string       `json:"backgroundAuditPolicy"`
	TamperAction               string       `json:"tamperAction"`
	UnsignedPolicy             string       `json:"unsignedPolicy"`
	SignatureTrustPolicy       string       `json:"signatureTrustPolicy"`
	AllowDevUnsigned           bool         `json:"allowDevUnsigned"`
	Extensions                 ExtensionMap `json:"extensions"`
}

// Lifecycle 描述插件生命周期和健康检查策略。
type Lifecycle struct {
	AutoStart           bool         `json:"autoStart"`
	StartTimeoutMs      int          `json:"startTimeoutMs"`
	RegisterTimeoutMs   int          `json:"registerTimeoutMs"`
	ReadyTimeoutMs      int          `json:"readyTimeoutMs"`
	StopTimeoutMs       int          `json:"stopTimeoutMs"`
	HeartbeatIntervalMs int          `json:"heartbeatIntervalMs"`
	HealthStrategy      string       `json:"healthStrategy"`
	ReloadPolicy        string       `json:"reloadPolicy"`
	UpgradePolicy       string       `json:"upgradePolicy"`
	RollbackPolicy      string       `json:"rollbackPolicy"`
	DrainPolicy         string       `json:"drainPolicy"`
	CrashBackoff        string       `json:"crashBackoff"`
	Extensions          ExtensionMap `json:"extensions"`
}

// Dependencies 描述插件依赖和冲突关系。
type Dependencies struct {
	Plugins               []Dependency `json:"plugins"`
	HostCapabilities      []string     `json:"hostCapabilities"`
	ExternalServices      []string     `json:"externalServices"`
	SystemRequirements    []string     `json:"systemRequirements"`
	Conflicts             []string     `json:"conflicts"`
	MigrationDependencies []string     `json:"migrationDependencies"`
	Extensions            ExtensionMap `json:"extensions"`
}

// Dependency 描述单个插件依赖。
type Dependency struct {
	ID         string       `json:"id"`
	Constraint string       `json:"constraint"`
	Optional   bool         `json:"optional"`
	Extensions ExtensionMap `json:"extensions"`
}

// Capabilities 是插件能力声明的分层集合。
type Capabilities struct {
	Backend       []string     `json:"backend"`
	Frontend      []string     `json:"frontend"`
	Data          []string     `json:"data"`
	Host          []string     `json:"host"`
	Security      []string     `json:"security"`
	Observability []string     `json:"observability"`
	Experimental  []string     `json:"experimental"`
	Extensions    ExtensionMap `json:"extensions"`
}

// Backend 描述插件后端能力。
type Backend struct {
	EntryExecutable   string                 `json:"entryExecutable"`
	ControlProtocol   string                 `json:"controlProtocol"`
	DataProtocol      string                 `json:"dataProtocol"`
	Routes            []RouteDeclaration     `json:"routes"`
	Webhooks          []RouteDeclaration     `json:"webhooks"`
	Events            []EventDeclaration     `json:"events"`
	Jobs              []JobDeclaration       `json:"jobs"`
	Consumers         []EventDeclaration     `json:"consumers"`
	Migrations        []MigrationDeclaration `json:"migrations"`
	SettingsRef       string                 `json:"settingsRef"`
	ConfigContracts   []string               `json:"configContracts"`
	ServiceContracts  []string               `json:"serviceContracts"`
	RPCServices       []string               `json:"rpcServices"`
	TransactionPolicy string                 `json:"transactionPolicy"`
	StoragePolicy     string                 `json:"storagePolicy"`
	DefaultTimeouts   map[string]int         `json:"defaultTimeouts"`
	Extensions        ExtensionMap           `json:"extensions"`
}

// Frontend 描述插件前端挂载能力。
type Frontend struct {
	Enabled              bool                    `json:"enabled"`
	ManifestRef          string                  `json:"manifestRef"`
	RenderMode           string                  `json:"renderMode"`
	MountAreas           []string                `json:"mountAreas"`
	Pages                []PageDeclaration       `json:"pages"`
	Menus                []MenuDeclaration       `json:"menus"`
	Forms                []FormDeclaration       `json:"forms"`
	Views                []ViewDeclaration       `json:"views"`
	SettingsPages        []PageDeclaration       `json:"settingsPages"`
	WorkspaceHints       WorkspaceHints          `json:"workspaceHints"`
	RenderModesSupported []string                `json:"renderModesSupported"`
	BundleEntries        []FrontendEntry         `json:"bundleEntries"`
	IframeEntries        []FrontendEntry         `json:"iframeEntries"`
	UIDependencyPolicy   string                  `json:"uiDependencyPolicy"`
	Permissions          []PermissionDeclaration `json:"permissions"`
	Extensions           ExtensionMap            `json:"extensions"`
}

// UIContribution 描述主题和 UI 扩展能力。
type UIContribution struct {
	Enabled               bool              `json:"enabled"`
	UIKind                string            `json:"uiKind"`
	ThemeScope            string            `json:"themeScope"`
	TokenExtensions       map[string]string `json:"tokenExtensions"`
	ComponentOverridesRef string            `json:"componentOverridesRef"`
	LayoutSkinRef         string            `json:"layoutSkinRef"`
	IconPackRef           string            `json:"iconPackRef"`
	ActivationPolicy      string            `json:"activationPolicy"`
	Extensions            ExtensionMap      `json:"extensions"`
}

// Observability 描述插件可观测性能力。
type Observability struct {
	HealthProbes      []string     `json:"healthProbes"`
	Metrics           []string     `json:"metrics"`
	LogChannels       []string     `json:"logChannels"`
	AuditEvents       []string     `json:"auditEvents"`
	DiagnosticBundles []string     `json:"diagnosticBundles"`
	TracingTags       []string     `json:"tracingTags"`
	SupportCommands   []string     `json:"supportCommands"`
	Extensions        ExtensionMap `json:"extensions"`
}

// Operations 描述安装、升级和维护钩子。
type Operations struct {
	InstallHooks     []string     `json:"installHooks"`
	EnableHooks      []string     `json:"enableHooks"`
	DisableHooks     []string     `json:"disableHooks"`
	UpgradeHooks     []string     `json:"upgradeHooks"`
	RollbackHooks    []string     `json:"rollbackHooks"`
	UninstallHooks   []string     `json:"uninstallHooks"`
	MaintenanceTasks []string     `json:"maintenanceTasks"`
	Extensions       ExtensionMap `json:"extensions"`
}

// PermissionDeclaration 描述宿主或插件权限点。
type PermissionDeclaration struct {
	Key               string       `json:"key"`
	Title             string       `json:"title"`
	Description       string       `json:"description"`
	Scope             string       `json:"scope"`
	Namespace         string       `json:"namespace"`
	Kind              string       `json:"kind"`
	RiskLevel         string       `json:"riskLevel"`
	DefaultVisibility string       `json:"defaultVisibility"`
	Dependencies      []string     `json:"dependencies"`
	Extensions        ExtensionMap `json:"extensions"`
}

// RouteDeclaration 描述插件可由宿主代理的路由能力。
type RouteDeclaration struct {
	ID               string       `json:"id"`
	Method           string       `json:"method"`
	Path             string       `json:"path"`
	HandlerKey       string       `json:"handlerKey"`
	Scope            string       `json:"scope"`
	AuthPolicy       string       `json:"authPolicy"`
	PermissionGuard  string       `json:"permissionGuard"`
	TimeoutMs        int          `json:"timeoutMs"`
	RateLimitProfile string       `json:"rateLimitProfile"`
	Version          string       `json:"version"`
	Extensions       ExtensionMap `json:"extensions"`
}

// EventDeclaration 描述事件发布或订阅能力。
type EventDeclaration struct {
	EventType       string       `json:"eventType"`
	EventVersion    string       `json:"eventVersion"`
	HandlerKey      string       `json:"handlerKey"`
	DeliveryMode    string       `json:"deliveryMode"`
	Ordering        string       `json:"ordering"`
	RetryPolicy     string       `json:"retryPolicy"`
	RequestedFields []string     `json:"requestedFields"`
	RiskLevel       string       `json:"riskLevel"`
	Extensions      ExtensionMap `json:"extensions"`
}

// JobDeclaration 描述异步任务能力。
type JobDeclaration struct {
	ID                string       `json:"id"`
	JobType           string       `json:"jobType"`
	ScheduleType      string       `json:"scheduleType"`
	ConcurrencyPolicy string       `json:"concurrencyPolicy"`
	TimeoutMs         int          `json:"timeoutMs"`
	RetryPolicy       string       `json:"retryPolicy"`
	Extensions        ExtensionMap `json:"extensions"`
}

// MigrationDeclaration 描述插件迁移声明。
type MigrationDeclaration struct {
	ID           string       `json:"id"`
	Version      string       `json:"version"`
	Direction    string       `json:"direction"`
	Path         string       `json:"path"`
	Checksum     string       `json:"checksum"`
	Dependencies []string     `json:"dependencies"`
	Extensions   ExtensionMap `json:"extensions"`
}

// DatabaseDeclaration 描述插件自己的数据库表声明。
// 宿主只接受 manifest 声明并登记后的插件表，规范见 Program/docs/Plugin_Development_Manual_CN/03-database-development.md。
type DatabaseDeclaration struct {
	Namespace   string                     `json:"namespace"`
	StorageMode string                     `json:"storageMode"`
	Tables      []DatabaseTableDeclaration `json:"tables"`
	Extensions  ExtensionMap               `json:"extensions"`
}

// DatabaseTableDeclaration 描述插件声明的一张业务表。
type DatabaseTableDeclaration struct {
	TableKey        string                         `json:"tableKey"`
	PhysicalName    string                         `json:"physicalName"`
	TableKind       string                         `json:"tableKind"`
	SchemaVersion   string                         `json:"schemaVersion"`
	SchemaChecksum  string                         `json:"schemaChecksum"`
	Description     string                         `json:"description"`
	Sensitivity     string                         `json:"sensitivity"`
	CreatePolicy    string                         `json:"createPolicy"`
	DropPolicy      string                         `json:"dropPolicy"`
	BackupPolicy    string                         `json:"backupPolicy"`
	RetentionPolicy string                         `json:"retentionPolicy"`
	Columns         []DatabaseColumnDeclaration    `json:"columns"`
	Indexes         []DatabaseIndexDeclaration     `json:"indexes"`
	Relations       []DatabaseRelationDeclaration  `json:"relations"`
	Operations      []DatabaseOperationDeclaration `json:"operations"`
	Extensions      ExtensionMap                   `json:"extensions"`
}

// DatabaseColumnDeclaration 描述插件表字段。
type DatabaseColumnDeclaration struct {
	ColumnKey       string       `json:"columnKey"`
	ColumnName      string       `json:"columnName"`
	DBType          string       `json:"dbType"`
	LogicalType     string       `json:"logicalType"`
	Nullable        bool         `json:"nullable"`
	DefaultValue    any          `json:"defaultValue"`
	PrimaryKey      bool         `json:"primaryKey"`
	AutoIncrement   bool         `json:"autoIncrement"`
	Unique          bool         `json:"unique"`
	Indexed         bool         `json:"indexed"`
	Encrypted       bool         `json:"encrypted"`
	Secret          bool         `json:"secret"`
	ReferenceType   string       `json:"referenceType"`
	ReferenceTarget string       `json:"referenceTarget"`
	Description     string       `json:"description"`
	Extensions      ExtensionMap `json:"extensions"`
}

// DatabaseIndexDeclaration 描述插件表索引。
type DatabaseIndexDeclaration struct {
	IndexKey   string       `json:"indexKey"`
	IndexName  string       `json:"indexName"`
	Columns    []string     `json:"columns"`
	Unique     bool         `json:"unique"`
	Extensions ExtensionMap `json:"extensions"`
}

// DatabaseRelationDeclaration 描述插件表到宿主资源或插件资源的逻辑关系。
type DatabaseRelationDeclaration struct {
	RelationKey        string       `json:"relationKey"`
	LocalColumn        string       `json:"localColumn"`
	TargetResourceType string       `json:"targetResourceType"`
	TargetKey          string       `json:"targetKey"`
	RelationType       string       `json:"relationType"`
	Required           bool         `json:"required"`
	OnDeletePolicy     string       `json:"onDeletePolicy"`
	Extensions         ExtensionMap `json:"extensions"`
}

// DatabaseOperationDeclaration 描述插件表需要宿主显式执行的结构操作。
type DatabaseOperationDeclaration struct {
	OperationID    string       `json:"operationId"`
	OperationType  string       `json:"operationType"`
	Path           string       `json:"path"`
	Checksum       string       `json:"checksum"`
	RequiresReview bool         `json:"requiresReview"`
	Extensions     ExtensionMap `json:"extensions"`
}

// PageDeclaration 描述前端页面。
type PageDeclaration struct {
	ID                string       `json:"id"`
	Area              string       `json:"area"`
	Path              string       `json:"path"`
	Title             string       `json:"title"`
	ViewID            string       `json:"viewId"`
	RenderMode        string       `json:"renderMode"`
	PermissionKeys    []string     `json:"permissionKeys"`
	AllowDirectAccess bool         `json:"allowDirectAccess"`
	Visible           bool         `json:"visible"`
	Extensions        ExtensionMap `json:"extensions"`
}

// MenuDeclaration 描述菜单入口。
type MenuDeclaration struct {
	ID                string       `json:"id"`
	TargetPageID      string       `json:"targetPageId"`
	Title             string       `json:"title"`
	Icon              string       `json:"icon"`
	DefaultGroup      string       `json:"defaultGroup"`
	Order             int          `json:"order"`
	Visible           bool         `json:"visible"`
	AllowWorkspacePin bool         `json:"allowWorkspacePin"`
	PermissionKeys    []string     `json:"permissionKeys"`
	Extensions        ExtensionMap `json:"extensions"`
}

// FormDeclaration 描述宿主渲染表单。
type FormDeclaration struct {
	ID              string        `json:"id"`
	Target          string        `json:"target"`
	Fields          []FieldSchema `json:"fields"`
	Layout          string        `json:"layout"`
	SubmitAction    string        `json:"submitAction"`
	SuccessBehavior string        `json:"successBehavior"`
	Extensions      ExtensionMap  `json:"extensions"`
}

// ViewDeclaration 描述宿主渲染视图。
type ViewDeclaration struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Schema     map[string]any `json:"schema"`
	Extensions ExtensionMap   `json:"extensions"`
}

// FieldSchema 描述配置或表单字段。
type FieldSchema struct {
	ID             string         `json:"id"`
	Key            string         `json:"key"`
	Label          string         `json:"label"`
	Description    string         `json:"description"`
	Type           string         `json:"type"`
	Required       bool           `json:"required"`
	Nullable       bool           `json:"nullable"`
	Secret         bool           `json:"secret"`
	Default        any            `json:"default"`
	EnumOptions    []string       `json:"enumOptions"`
	Validation     map[string]any `json:"validation"`
	UI             map[string]any `json:"ui"`
	BackendBinding string         `json:"backendBinding"`
	Scope          string         `json:"scope"`
	ReloadPolicy   string         `json:"reloadPolicy"`
	Visibility     string         `json:"visibility"`
	DependsOn      []string       `json:"dependsOn"`
	Deprecated     bool           `json:"deprecated"`
	Examples       []string       `json:"examples"`
	Extensions     ExtensionMap   `json:"extensions"`
}

// FrontendEntry 描述 bundle 或 iframe 入口。
type FrontendEntry struct {
	ID            string       `json:"id"`
	Entry         string       `json:"entry"`
	Integrity     string       `json:"integrity"`
	SandboxPolicy string       `json:"sandboxPolicy"`
	Extensions    ExtensionMap `json:"extensions"`
}

// WorkspaceHints 描述后台工作区推荐组织方式。
type WorkspaceHints struct {
	RecommendedGroups    []string     `json:"recommendedGroups"`
	RecommendedShortcuts []string     `json:"recommendedShortcuts"`
	DefaultEntry         string       `json:"defaultEntry"`
	AllowCustomGrouping  bool         `json:"allowCustomGrouping"`
	Extensions           ExtensionMap `json:"extensions"`
}

// RuntimePlugin 是请求期内存能力表的插件视图。
type RuntimePlugin struct {
	SessionID               string       `json:"sessionId"`
	PluginID                string       `json:"pluginId"`
	InstallID               string       `json:"installId"`
	InstanceID              string       `json:"instanceId"`
	PID                     int          `json:"pid"`
	State                   string       `json:"state"`
	TrafficEnabled          bool         `json:"trafficEnabled"`
	SelectedProtocolVersion string       `json:"selectedProtocolVersion"`
	ControlTransport        string       `json:"controlTransport"`
	DataTransport           string       `json:"dataTransport"`
	ChannelEndpoint         string       `json:"channelEndpoint"`
	CapabilityHashDeclared  string       `json:"capabilityHashDeclared"`
	CapabilityHashEffective string       `json:"capabilityHashEffective"`
	CurrentManifestHash     string       `json:"currentManifestHash"`
	CurrentBinaryHash       string       `json:"currentBinaryHash"`
	TrustLevel              string       `json:"trustLevel"`
	IntegrityState          string       `json:"integrityState"`
	ConfigVersion           int          `json:"configVersion"`
	StartedAt               time.Time    `json:"startedAt"`
	ReadyAt                 time.Time    `json:"readyAt"`
	LastHeartbeatAt         time.Time    `json:"lastHeartbeatAt"`
	LoadSummary             string       `json:"loadSummary"`
	RecentErrorCount        int          `json:"recentErrorCount"`
	RecentLatencyMs         int          `json:"recentLatencyMs"`
	Extensions              ExtensionMap `json:"extensions"`
}

// CapabilitySnapshot 是插件握手后提交的运行时能力快照。
type CapabilitySnapshot struct {
	SessionID       string                  `json:"sessionId"`
	Routes          []RouteDeclaration      `json:"routes"`
	Events          []EventDeclaration      `json:"events"`
	Jobs            []JobDeclaration        `json:"jobs"`
	Permissions     []PermissionDeclaration `json:"permissions"`
	SettingsSchema  ConfigSchema            `json:"settingsSchema"`
	FrontendRefs    []PageDeclaration       `json:"frontendRefs"`
	CapabilityHash  string                  `json:"capabilityHash"`
	SnapshotVersion string                  `json:"snapshotVersion"`
	Extensions      ExtensionMap            `json:"extensions"`
}

// ConfigSchema 描述插件配置结构。
type ConfigSchema struct {
	SchemaVersion    string          `json:"schemaVersion"`
	PluginID         string          `json:"pluginId"`
	ConfigVersion    string          `json:"configVersion"`
	Storage          map[string]any  `json:"storage"`
	Scopes           []string        `json:"scopes"`
	Sections         []ConfigSection `json:"sections"`
	DefaultsRef      string          `json:"defaultsRef"`
	SecretPolicies   []string        `json:"secretPolicies"`
	ValidationRules  []string        `json:"validationRules"`
	ReloadPolicies   []string        `json:"reloadPolicies"`
	AuditPolicy      string          `json:"auditPolicy"`
	MigrationRules   []string        `json:"migrationRules"`
	PermissionGuards []string        `json:"permissionGuards"`
	UIHints          map[string]any  `json:"uiHints"`
	Extensions       ExtensionMap    `json:"extensions"`
}

// ConfigSection 描述配置页分组。
type ConfigSection struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Scope       string        `json:"scope"`
	Order       int           `json:"order"`
	Permission  string        `json:"permission"`
	Layout      string        `json:"layout"`
	Fields      []FieldSchema `json:"fields"`
	Extensions  ExtensionMap  `json:"extensions"`
}

// HostLaunchContext 是宿主拉起插件进程时下发的启动上下文。
type HostLaunchContext struct {
	PluginID                string       `json:"pluginId"`
	PluginVersionExpected   string       `json:"pluginVersionExpected"`
	InstallID               string       `json:"installId"`
	InstanceID              string       `json:"instanceId"`
	HandshakeToken          string       `json:"handshakeToken"`
	HostProtocolVersion     string       `json:"hostProtocolVersion"`
	ManifestHash            string       `json:"manifestHash"`
	ApprovedPackageHash     string       `json:"approvedPackageHash"`
	ApprovedBinaryHash      string       `json:"approvedBinaryHash"`
	ConfigHash              string       `json:"configHash"`
	IntegrityPolicyVersion  string       `json:"integrityPolicyVersion"`
	TrustLevelExpected      string       `json:"trustLevelExpected"`
	SignatureStatusExpected string       `json:"signatureStatusExpected"`
	ControlTransport        string       `json:"controlTransport"`
	DataTransportOffer      string       `json:"dataTransportOffer"`
	StartupDeadlineMs       int          `json:"startupDeadlineMs"`
	RegisterDeadlineMs      int          `json:"registerDeadlineMs"`
	HeartbeatIntervalMs     int          `json:"heartbeatIntervalMs"`
	HostRuntimeMode         string       `json:"hostRuntimeMode"`
	Extensions              ExtensionMap `json:"extensions"`
}

// IdentityContext 是宿主传递给插件的最小请求身份上下文。
type IdentityContext struct {
	SubjectType     string       `json:"subjectType"`
	SubjectID       string       `json:"subjectId"`
	SessionID       string       `json:"sessionId"`
	IsAuthenticated bool         `json:"isAuthenticated"`
	AuthMethod      string       `json:"authMethod"`
	AuthStrength    string       `json:"authStrength"`
	RoleKeys        []string     `json:"roleKeys"`
	GrantedScopes   []string     `json:"grantedScopes"`
	TenantID        string       `json:"tenantId"`
	RequestID       string       `json:"requestId"`
	TraceID         string       `json:"traceId"`
	ClientIP        string       `json:"clientIp"`
	UserAgent       string       `json:"userAgent"`
	Locale          string       `json:"locale"`
	Timezone        string       `json:"timezone"`
	Extensions      ExtensionMap `json:"extensions"`
}

// EventEnvelope 是宿主事件总线的统一事件外壳。
type EventEnvelope struct {
	EventID      string       `json:"eventId"`
	EventType    string       `json:"eventType"`
	EventVersion string       `json:"eventVersion"`
	OccurredAt   time.Time    `json:"occurredAt"`
	ProducerType string       `json:"producerType"`
	ProducerID   string       `json:"producerId"`
	RequestID    string       `json:"requestId"`
	SubjectID    string       `json:"subjectId"`
	ResourceType string       `json:"resourceType"`
	ResourceID   string       `json:"resourceId"`
	Payload      any          `json:"payload"`
	Extensions   ExtensionMap `json:"extensions"`
}

// JobRequest 是异步任务调度接口的统一请求。
type JobRequest struct {
	JobType      string       `json:"jobType"`
	ResourceType string       `json:"resourceType"`
	ResourceID   string       `json:"resourceId"`
	Priority     int          `json:"priority"`
	RunAt        time.Time    `json:"runAt"`
	MaxRetries   int          `json:"maxRetries"`
	Payload      any          `json:"payload"`
	Extensions   ExtensionMap `json:"extensions"`
}

// Registry 定义宿主插件注册中心能力。
type Registry interface {
	RegisterManifest(ctx context.Context, manifest Manifest) error
	GetManifest(pluginID string) (Manifest, bool)
	ListManifests() []Manifest
	GetRuntime(pluginID string) (RuntimePlugin, bool)
	ListRuntimes() []RuntimePlugin
	FrontendContribution() FrontendContribution
	ThemeContribution() []UIContribution
}

// EventBus 定义宿主事件总线最小接口。
type EventBus interface {
	Publish(ctx context.Context, event EventEnvelope) error
	Subscribe(eventType string, handler EventHandler) error
}

// EventHandler 是事件处理器。
type EventHandler func(context.Context, EventEnvelope) error

// MigrationRegistry 定义插件迁移注册接口。
type MigrationRegistry interface {
	Register(pluginID string, migrations []MigrationDeclaration) error
	List(pluginID string) []MigrationDeclaration
}

// ConfigRegistry 定义插件配置 schema 注册接口。
type ConfigRegistry interface {
	Register(pluginID string, schema ConfigSchema) error
	Get(pluginID string) (ConfigSchema, bool)
}

// ThemeRegistry 定义主题插件注册接口。
type ThemeRegistry interface {
	Register(pluginID string, contribution UIContribution) error
	List() []UIContribution
}

// PermissionRegistry 定义插件权限注册接口。
type PermissionRegistry interface {
	Register(pluginID string, permissions []PermissionDeclaration) error
	List() []PermissionDeclaration
}

// FrontendContribution 是前端可读取的菜单、页面、视图聚合结果。
type FrontendContribution struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Pages           []PageDeclaration `json:"pages"`
	Menus           []MenuDeclaration `json:"menus"`
	Forms           []FormDeclaration `json:"forms"`
	Views           []ViewDeclaration `json:"views"`
	Themes          []UIContribution  `json:"themes"`
	Extensions      ExtensionMap      `json:"extensions"`
}
