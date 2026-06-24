package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PermissionOwnerHost   = "host"
	PermissionOwnerPlugin = "plugin"

	DefaultPermissionRiskLevel = "normal"
	DefaultGrantPolicyManual   = "manual"
)

// PermissionDefinitionInput 是权限定义注册输入。
type PermissionDefinitionInput struct {
	PermissionCode     string
	OwnerType          string
	OwnerPluginID      string
	RiskLevel          string
	GroupKey           string
	Name               string
	Description        string
	DefaultGrantPolicy string
	Extensions         any
}

// ResourceScopeDefinitionInput 是数据范围定义注册输入。
type ResourceScopeDefinitionInput struct {
	ResourceType  string
	ScopeType     string
	OwnerPluginID string
	Name          string
	Description   string
	Extensions    any
}

// AuditInput 是审计日志输入。
type AuditInput struct {
	ActorSubjectID string
	Action         string
	ResourceType   string
	ResourceID     string
	RiskLevel      string
	IP             string
	UserAgent      string
	Payload        any
}

// EventInput 是统一事件日志输入。
type EventInput struct {
	EventID       string
	EventType     string
	SourceType    string
	SourceID      string
	OwnerPluginID string
	Payload       any
}

// SystemJobInput 是统一任务创建输入。
type SystemJobInput struct {
	JobID         string
	JobType       string
	OwnerPluginID string
	RunAt         time.Time
	MaxAttempts   int
	Payload       any
}

// PluginConfigValueInput 是插件配置保存输入。
type PluginConfigValueInput struct {
	PluginID      string
	ConfigKey     string
	Value         any
	SecretJSON    string
	UpdatedBy     string
	ChangeSummary string
}

// PluginVersionInput 是插件版本登记输入。
type PluginVersionInput struct {
	PluginID     string
	Version      string
	ManifestHash string
	PackageHash  string
	BinaryHash   string
	InstallPath  string
	Status       string
}

// GovernanceService 负责权限、数据范围、事件、任务和审计治理。
type GovernanceService struct {
	repo *repository.Repository
}

func NewGovernanceService(repo *repository.Repository) *GovernanceService {
	return &GovernanceService{repo: repo}
}

// SyncHostPermissions 将当前宿主静态权限注册到权限定义表。
func (s *GovernanceService) SyncHostPermissions(permissions []model.Permission) error {
	inputs := make([]PermissionDefinitionInput, 0, len(permissions))
	for _, permission := range permissions {
		inputs = append(inputs, PermissionDefinitionInput{
			PermissionCode:     permission.Code,
			OwnerType:          PermissionOwnerHost,
			RiskLevel:          DefaultPermissionRiskLevel,
			GroupKey:           permission.Group,
			Name:               permission.Name,
			Description:        permission.Description,
			DefaultGrantPolicy: DefaultGrantPolicyManual,
		})
	}
	return s.RegisterPermissions(inputs)
}

// RegisterPermissions 批量注册权限定义。
func (s *GovernanceService) RegisterPermissions(inputs []PermissionDefinitionInput) error {
	if s == nil || s.repo == nil {
		return errors.New("治理服务未初始化")
	}
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		for _, input := range inputs {
			code := strings.TrimSpace(input.PermissionCode)
			if code == "" {
				return errors.New("权限代码不能为空")
			}
			var existing model.PermissionDefinition
			err := tx.Where("permission_code = ?", code).First(&existing).Error
			record := model.PermissionDefinition{
				PermissionCode:     code,
				OwnerType:          defaultText(input.OwnerType, PermissionOwnerHost),
				OwnerPluginID:      strings.TrimSpace(input.OwnerPluginID),
				RiskLevel:          defaultText(input.RiskLevel, DefaultPermissionRiskLevel),
				GroupKey:           strings.TrimSpace(input.GroupKey),
				Name:               strings.TrimSpace(input.Name),
				Description:        strings.TrimSpace(input.Description),
				DefaultGrantPolicy: defaultText(input.DefaultGrantPolicy, DefaultGrantPolicyManual),
				Status:             "active",
				ExtensionsJSON:     mustJSON(input.Extensions),
			}
			if err == nil {
				record.ID = existing.ID
				if err := tx.Save(&record).Error; err != nil {
					return err
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GovernanceService) ListPermissionDefinitions() ([]model.PermissionDefinition, error) {
	var records []model.PermissionDefinition
	err := s.repo.GetDB().Order("owner_type ASC, owner_plugin_id ASC, permission_code ASC").Find(&records).Error
	return records, err
}

func (s *GovernanceService) ListPluginPermissionDefinitions(pluginID string) ([]model.PermissionDefinition, error) {
	var records []model.PermissionDefinition
	err := s.repo.GetDB().
		Where("owner_type = ? AND owner_plugin_id = ?", PermissionOwnerPlugin, strings.TrimSpace(pluginID)).
		Order("permission_code ASC").
		Find(&records).Error
	return records, err
}

func (s *GovernanceService) UpsertPluginVersion(input PluginVersionInput) error {
	pluginID := strings.TrimSpace(input.PluginID)
	version := strings.TrimSpace(input.Version)
	if pluginID == "" || version == "" {
		return errors.New("插件ID和版本不能为空")
	}
	record := model.PluginVersion{
		PluginID:     pluginID,
		Version:      version,
		ManifestHash: strings.TrimSpace(input.ManifestHash),
		PackageHash:  strings.TrimSpace(input.PackageHash),
		BinaryHash:   strings.TrimSpace(input.BinaryHash),
		InstallPath:  strings.TrimSpace(input.InstallPath),
		Status:       defaultText(input.Status, "installed"),
		InstalledAt:  time.Now(),
	}
	var existing model.PluginVersion
	err := s.repo.GetDB().Where("plugin_id = ? AND version = ?", pluginID, version).First(&existing).Error
	if err == nil {
		record.ID = existing.ID
		record.InstalledAt = existing.InstalledAt
		return s.repo.GetDB().Save(&record).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) EnsurePluginTrustRecord(pluginID, version, riskSummary string, requiresReview bool) error {
	pluginID = strings.TrimSpace(pluginID)
	version = strings.TrimSpace(version)
	if pluginID == "" || version == "" {
		return errors.New("插件ID和版本不能为空")
	}
	var existing model.PluginTrustRecord
	err := s.repo.GetDB().Where("plugin_id = ? AND version = ?", pluginID, version).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	trustLevel := "local-approved"
	if requiresReview {
		trustLevel = "pending-review"
	}
	record := model.PluginTrustRecord{
		PluginID:        pluginID,
		Version:         version,
		TrustLevel:      trustLevel,
		SignatureStatus: "unknown",
		RiskSummary:     riskSummary,
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) ApprovePluginTrust(pluginID, version, approvedBy string) error {
	now := time.Now()
	return s.repo.GetDB().Model(&model.PluginTrustRecord{}).
		Where("plugin_id = ? AND version = ?", strings.TrimSpace(pluginID), strings.TrimSpace(version)).
		Updates(map[string]any{
			"trust_level": "local-approved",
			"approved_by": strings.TrimSpace(approvedBy),
			"approved_at": &now,
		}).Error
}

func (s *GovernanceService) ListPluginTrustRecords(pluginID string) ([]model.PluginTrustRecord, error) {
	var records []model.PluginTrustRecord
	query := s.repo.GetDB().Model(&model.PluginTrustRecord{})
	if strings.TrimSpace(pluginID) != "" {
		query = query.Where("plugin_id = ?", strings.TrimSpace(pluginID))
	}
	err := query.Order("plugin_id ASC, version DESC").Find(&records).Error
	return records, err
}

func (s *GovernanceService) AssertPluginTrusted(pluginID, version string) error {
	var record model.PluginTrustRecord
	err := s.repo.GetDB().Where("plugin_id = ? AND version = ?", strings.TrimSpace(pluginID), strings.TrimSpace(version)).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("插件版本未完成信任登记，不能启用")
		}
		return err
	}
	switch record.TrustLevel {
	case "local-approved", "approved":
		return nil
	default:
		return errors.New("插件版本未批准，不能启用")
	}
}

func (s *GovernanceService) GrantRolePermission(roleID uint, permissionCode, grantedBy string) error {
	if roleID == 0 || strings.TrimSpace(permissionCode) == "" {
		return errors.New("角色和权限不能为空")
	}
	record := model.RolePermissionGrant{
		RoleID:             roleID,
		PermissionCode:     permissionCode,
		GrantedBySubjectID: grantedBy,
		GrantedAt:          time.Now(),
	}
	var existing model.RolePermissionGrant
	err := s.repo.GetDB().Where("role_id = ? AND permission_code = ?", roleID, permissionCode).First(&existing).Error
	if err == nil {
		record.ID = existing.ID
		return s.repo.GetDB().Save(&record).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) SyncRolePermissionGrants(roleID uint, permissions []string, grantedBy string) error {
	if roleID == 0 {
		return errors.New("角色不能为空")
	}
	now := time.Now()
	seen := make(map[string]bool, len(permissions))
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermissionGrant{}).Error; err != nil {
			return err
		}
		for _, permission := range permissions {
			code := strings.TrimSpace(permission)
			if code == "" || seen[code] {
				continue
			}
			seen[code] = true
			record := model.RolePermissionGrant{
				RoleID:             roleID,
				PermissionCode:     code,
				GrantedBySubjectID: grantedBy,
				GrantedAt:          now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GovernanceService) RegisterResourceScopes(inputs []ResourceScopeDefinitionInput) error {
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		for _, input := range inputs {
			if strings.TrimSpace(input.ResourceType) == "" || strings.TrimSpace(input.ScopeType) == "" {
				return errors.New("资源类型和范围类型不能为空")
			}
			record := model.ResourceScopeDefinition{
				ResourceType:   input.ResourceType,
				ScopeType:      input.ScopeType,
				OwnerPluginID:  input.OwnerPluginID,
				Name:           input.Name,
				Description:    input.Description,
				Status:         "active",
				ExtensionsJSON: mustJSON(input.Extensions),
			}
			var existing model.ResourceScopeDefinition
			err := tx.Where("resource_type = ? AND scope_type = ? AND owner_plugin_id = ?", input.ResourceType, input.ScopeType, input.OwnerPluginID).First(&existing).Error
			if err == nil {
				record.ID = existing.ID
				if err := tx.Save(&record).Error; err != nil {
					return err
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GovernanceService) GrantSubjectDataScope(subjectID, subjectType string, scope DataScope, grantedBy string) error {
	if strings.TrimSpace(subjectID) == "" || strings.TrimSpace(scope.ResourceType) == "" || strings.TrimSpace(scope.ScopeType) == "" {
		return errors.New("主体和数据范围不能为空")
	}
	record := model.SubjectDataScopeGrant{
		SubjectID:          subjectID,
		SubjectType:        subjectType,
		ResourceType:       scope.ResourceType,
		ScopeType:          scope.ScopeType,
		ScopeValue:         scope.ScopeValue,
		OwnerPluginID:      scope.OwnerPluginID,
		GrantedBySubjectID: grantedBy,
		GrantedAt:          time.Now(),
	}
	var existing model.SubjectDataScopeGrant
	err := s.repo.GetDB().Where(
		"subject_id = ? AND resource_type = ? AND scope_type = ? AND scope_value = ? AND owner_plugin_id = ?",
		record.SubjectID,
		record.ResourceType,
		record.ScopeType,
		record.ScopeValue,
		record.OwnerPluginID,
	).First(&existing).Error
	if err == nil {
		record.ID = existing.ID
		return s.repo.GetDB().Save(&record).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) ListSubjectDataScopes(subjectID string) ([]DataScope, error) {
	var grants []model.SubjectDataScopeGrant
	if err := s.repo.GetDB().Where("subject_id = ?", subjectID).Order("resource_type ASC, scope_type ASC").Find(&grants).Error; err != nil {
		return nil, err
	}
	scopes := make([]DataScope, 0, len(grants))
	for _, grant := range grants {
		scopes = append(scopes, DataScope{
			ResourceType:  grant.ResourceType,
			ScopeType:     grant.ScopeType,
			ScopeValue:    grant.ScopeValue,
			OwnerPluginID: grant.OwnerPluginID,
		})
	}
	return scopes, nil
}

func (s *GovernanceService) SyncPluginConfigSchemas(schemas []pluginapi.ConfigSchema) error {
	if len(schemas) == 0 {
		return nil
	}
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		for _, schema := range schemas {
			pluginID := strings.TrimSpace(schema.PluginID)
			if pluginID == "" {
				return errors.New("插件配置 schema 缺少插件ID")
			}
			configKey := "default"
			schemaJSON := mustJSON(schema)
			record := model.PluginConfigSchema{
				PluginID:      pluginID,
				ConfigKey:     configKey,
				SchemaVersion: defaultText(schema.SchemaVersion, "1.0.0"),
				SchemaJSON:    schemaJSON,
				Status:        "active",
			}
			var existing model.PluginConfigSchema
			err := tx.Where("plugin_id = ? AND config_key = ?", pluginID, configKey).First(&existing).Error
			if err == nil {
				record.ID = existing.ID
				if err := tx.Save(&record).Error; err != nil {
					return err
				}
				if err := ensurePluginDefaultConfigValue(tx, schema); err != nil {
					return err
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			if err := ensurePluginDefaultConfigValue(tx, schema); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GovernanceService) ListPluginConfigSchemas(pluginID string) ([]model.PluginConfigSchema, error) {
	var records []model.PluginConfigSchema
	query := s.repo.GetDB().Model(&model.PluginConfigSchema{})
	if strings.TrimSpace(pluginID) != "" {
		query = query.Where("plugin_id = ?", strings.TrimSpace(pluginID))
	}
	err := query.Order("plugin_id ASC, config_key ASC").Find(&records).Error
	return records, err
}

func (s *GovernanceService) SavePluginConfigValue(input PluginConfigValueInput) (*model.PluginConfigValue, error) {
	pluginID := strings.TrimSpace(input.PluginID)
	configKey := defaultText(input.ConfigKey, "default")
	if pluginID == "" {
		return nil, errors.New("插件ID不能为空")
	}
	valueJSON := mustJSON(input.Value)
	var saved model.PluginConfigValue
	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing model.PluginConfigValue
		err := tx.Where("plugin_id = ? AND config_key = ?", pluginID, configKey).First(&existing).Error
		revision := 1
		if err == nil {
			revision = existing.Revision + 1
			if strings.TrimSpace(input.SecretJSON) == "" {
				input.SecretJSON = existing.SecretJSON
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		valueDigest := sha256Hex(valueJSON + input.SecretJSON)

		record := model.PluginConfigValue{
			PluginID:   pluginID,
			ConfigKey:  configKey,
			ValueJSON:  valueJSON,
			SecretJSON: strings.TrimSpace(input.SecretJSON),
			Revision:   revision,
			UpdatedBy:  strings.TrimSpace(input.UpdatedBy),
		}
		if existing.ID != 0 {
			record.ID = existing.ID
			if err := tx.Save(&record).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&record).Error; err != nil {
			return err
		}

		revisionRecord := model.PluginConfigRevision{
			PluginID:      pluginID,
			ConfigKey:     configKey,
			Revision:      revision,
			ValueDigest:   valueDigest,
			SecretJSON:    strings.TrimSpace(input.SecretJSON),
			UpdatedBy:     strings.TrimSpace(input.UpdatedBy),
			ChangeSummary: strings.TrimSpace(input.ChangeSummary),
		}
		if err := tx.Create(&revisionRecord).Error; err != nil {
			return err
		}
		saved = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	if saved.SecretJSON != "" {
		saved.SecretJSON = "__saved__"
	}
	return &saved, nil
}

func (s *GovernanceService) ListPluginConfigValues(pluginID string) ([]model.PluginConfigValue, error) {
	var records []model.PluginConfigValue
	err := s.repo.GetDB().
		Where("plugin_id = ?", strings.TrimSpace(pluginID)).
		Order("config_key ASC").
		Find(&records).Error
	for i := range records {
		if records[i].SecretJSON != "" {
			records[i].SecretJSON = "__saved__"
		}
	}
	return records, err
}

func (s *GovernanceService) ListPluginConfigRevisions(pluginID string) ([]model.PluginConfigRevision, error) {
	var records []model.PluginConfigRevision
	err := s.repo.GetDB().
		Where("plugin_id = ?", strings.TrimSpace(pluginID)).
		Order("id DESC").
		Limit(200).
		Find(&records).Error
	for i := range records {
		if records[i].SecretJSON != "" {
			records[i].SecretJSON = "__saved__"
		}
	}
	return records, err
}

func ensurePluginDefaultConfigValue(tx *gorm.DB, schema pluginapi.ConfigSchema) error {
	pluginID := strings.TrimSpace(schema.PluginID)
	if pluginID == "" {
		return errors.New("插件配置 schema 缺少插件ID")
	}
	configKey := "default"
	defaultValues := defaultConfigValues(schema)

	var existing model.PluginConfigValue
	err := tx.Where("plugin_id = ? AND config_key = ?", pluginID, configKey).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		valueJSON := mustJSON(defaultValues)
		record := model.PluginConfigValue{
			PluginID:   pluginID,
			ConfigKey:  configKey,
			ValueJSON:  valueJSON,
			SecretJSON: "",
			Revision:   1,
			UpdatedBy:  "system:plugin-config-defaults",
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return tx.Create(&model.PluginConfigRevision{
			PluginID:      pluginID,
			ConfigKey:     configKey,
			Revision:      1,
			ValueDigest:   sha256Hex(valueJSON),
			UpdatedBy:     record.UpdatedBy,
			ChangeSummary: "插件配置 schema 首次同步时初始化默认配置",
		}).Error
	}

	values := map[string]any{}
	if strings.TrimSpace(existing.ValueJSON) != "" {
		_ = json.Unmarshal([]byte(existing.ValueJSON), &values)
	}
	changed := false
	for key, value := range defaultValues {
		if _, ok := values[key]; ok {
			continue
		}
		values[key] = value
		changed = true
	}
	if !changed {
		return nil
	}

	valueJSON := mustJSON(values)
	existing.ValueJSON = valueJSON
	existing.Revision++
	existing.UpdatedBy = "system:plugin-config-defaults"
	if err := tx.Save(&existing).Error; err != nil {
		return err
	}
	return tx.Create(&model.PluginConfigRevision{
		PluginID:      pluginID,
		ConfigKey:     configKey,
		Revision:      existing.Revision,
		ValueDigest:   sha256Hex(valueJSON + existing.SecretJSON),
		SecretJSON:    existing.SecretJSON,
		UpdatedBy:     existing.UpdatedBy,
		ChangeSummary: "插件配置 schema 同步时补齐新增默认字段",
	}).Error
}

func defaultConfigValues(schema pluginapi.ConfigSchema) map[string]any {
	values := map[string]any{}
	secretKeys := map[string]bool{}
	for _, key := range schema.SecretPolicies {
		key = strings.TrimSpace(key)
		if key != "" {
			secretKeys[key] = true
		}
	}
	for _, section := range schema.Sections {
		for _, field := range section.Fields {
			key := strings.TrimSpace(field.Key)
			if key == "" {
				key = strings.TrimSpace(field.ID)
			}
			if key == "" || field.Secret || secretKeys[key] || strings.EqualFold(field.Type, "secret") {
				continue
			}
			if field.Default != nil {
				values[key] = field.Default
				continue
			}
			values[key] = emptyConfigValue(field.Type)
		}
	}
	return values
}

func emptyConfigValue(fieldType string) any {
	switch strings.ToLower(strings.TrimSpace(fieldType)) {
	case "bool", "boolean":
		return false
	case "number", "integer", "int", "float":
		return 0
	case "array", "list", "multi-select", "multiselect":
		return []any{}
	case "object", "map", "json":
		return map[string]any{}
	default:
		return ""
	}
}

func (s *GovernanceService) RecordPluginStateEvent(pluginID, fromState, toState, eventType, reason, operatorSubjectID string) error {
	record := model.PluginStateEvent{
		PluginID:          strings.TrimSpace(pluginID),
		FromState:         strings.TrimSpace(fromState),
		ToState:           strings.TrimSpace(toState),
		EventType:         strings.TrimSpace(eventType),
		Reason:            strings.TrimSpace(reason),
		OperatorSubjectID: strings.TrimSpace(operatorSubjectID),
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) RecordPluginFault(pluginID, instanceID, faultType, reason, stackTrace string) error {
	record := model.PluginFaultLog{
		PluginID:    strings.TrimSpace(pluginID),
		InstanceID:  strings.TrimSpace(instanceID),
		FaultType:   strings.TrimSpace(faultType),
		FaultReason: strings.TrimSpace(reason),
		StackTrace:  stackTrace,
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) ListPluginStateEvents(pluginID string) ([]model.PluginStateEvent, error) {
	var records []model.PluginStateEvent
	err := s.repo.GetDB().
		Where("plugin_id = ?", strings.TrimSpace(pluginID)).
		Order("id DESC").
		Limit(200).
		Find(&records).Error
	return records, err
}

func (s *GovernanceService) ListPluginFaultLogs(pluginID string) ([]model.PluginFaultLog, error) {
	var records []model.PluginFaultLog
	err := s.repo.GetDB().
		Where("plugin_id = ?", strings.TrimSpace(pluginID)).
		Order("id DESC").
		Limit(200).
		Find(&records).Error
	return records, err
}

// DeletePluginGovernanceData 删除插件配置与治理数据。
func (s *GovernanceService) DeletePluginGovernanceData(pluginID string) error {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return errors.New("插件ID不能为空")
	}
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		var permissionCodes []string
		if err := tx.Model(&model.PermissionDefinition{}).
			Where("owner_type = ? AND owner_plugin_id = ?", PermissionOwnerPlugin, pluginID).
			Pluck("permission_code", &permissionCodes).Error; err != nil {
			return err
		}
		if len(permissionCodes) > 0 {
			if err := tx.Where("permission_code IN ?", permissionCodes).Delete(&model.RolePermissionGrant{}).Error; err != nil {
				return err
			}
			if err := tx.Where("permission_code IN ?", permissionCodes).Delete(&model.SubjectPermissionGrant{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginStateEvent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginFaultLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginTrustRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginVersion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_type = ? AND owner_plugin_id = ?", PermissionOwnerPlugin, pluginID).Delete(&model.PermissionDefinition{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_plugin_id = ?", pluginID).Delete(&model.ResourceScopeDefinition{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_plugin_id = ?", pluginID).Delete(&model.SubjectDataScopeGrant{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_plugin_id = ?", pluginID).Delete(&model.EventLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_plugin_id = ?", pluginID).Delete(&model.SystemJob{}).Error; err != nil {
			return err
		}
		if err := tx.Where("owner_plugin_id = ?", pluginID).Delete(&model.PluginJob{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginRuntimeSession{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeletePluginConfigData 删除插件配置数据。
func (s *GovernanceService) DeletePluginConfigData(pluginID string) error {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return errors.New("插件ID不能为空")
	}
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginConfigRevision{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginConfigValue{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginConfigSchema{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *GovernanceService) RecordAudit(input AuditInput) error {
	payloadJSON := mustJSON(input.Payload)
	record := model.AuditLog{
		ActorSubjectID: strings.TrimSpace(input.ActorSubjectID),
		Action:         strings.TrimSpace(input.Action),
		ResourceType:   strings.TrimSpace(input.ResourceType),
		ResourceID:     strings.TrimSpace(input.ResourceID),
		RiskLevel:      defaultText(input.RiskLevel, DefaultPermissionRiskLevel),
		IP:             input.IP,
		UserAgent:      input.UserAgent,
		PayloadDigest:  sha256Hex(payloadJSON),
		PayloadJSON:    payloadJSON,
		CreatedAt:      time.Now(),
	}
	return s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) RecordEvent(input EventInput) (string, error) {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		eventID = "evt_" + uuid.NewString()
	}
	record := model.EventLog{
		EventID:       eventID,
		EventType:     input.EventType,
		SourceType:    input.SourceType,
		SourceID:      input.SourceID,
		OwnerPluginID: input.OwnerPluginID,
		PayloadJSON:   mustJSON(input.Payload),
		Status:        "recorded",
	}
	return eventID, s.repo.GetDB().Create(&record).Error
}

func (s *GovernanceService) CreateSystemJob(input SystemJobInput) (string, error) {
	jobID := strings.TrimSpace(input.JobID)
	if jobID == "" {
		jobID = "job_" + uuid.NewString()
	}
	if strings.TrimSpace(input.JobType) == "" {
		return "", errors.New("任务类型不能为空")
	}
	runAt := input.RunAt
	if runAt.IsZero() {
		runAt = time.Now()
	}
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	record := model.SystemJob{
		JobID:         jobID,
		JobType:       input.JobType,
		OwnerPluginID: input.OwnerPluginID,
		Status:        "pending",
		RunAt:         runAt,
		MaxAttempts:   maxAttempts,
		PayloadJSON:   mustJSON(input.Payload),
	}
	return jobID, s.repo.GetDB().Create(&record).Error
}

func mustJSON(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
