package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PluginRuntimeStateStarting    = pluginapi.PluginRuntimeStarting
	PluginRuntimeStateHandshaking = pluginapi.PluginRuntimeHandshaking
	PluginRuntimeStateReady       = pluginapi.PluginRuntimeReady
	PluginRuntimeStateStopped     = pluginapi.PluginRuntimeStopped
	PluginRuntimeStateCrashed     = pluginapi.PluginRuntimeCrashed
	PluginRuntimeStateQuarantined = pluginapi.PluginStateQuarantined
)

// ResolvedPluginBinary 是当前平台选中的插件二进制。
type ResolvedPluginBinary struct {
	PluginID   string
	Version    string
	ReleaseDir string
	Path       string
	SHA256     string
	Platform   string
	Arch       string
}

// PluginInvocationRequest 是插件调用网关输入。
type PluginInvocationRequest struct {
	PluginID  string
	RouteID   string
	Timeout   time.Duration
	Payload   any
	SubjectID string
}

// PluginInvocationResult 是插件调用网关输出。
type PluginInvocationResult struct {
	Accepted bool
	Message  string
}

type pluginProcess struct {
	cmd        *exec.Cmd
	instanceID string
}

// PluginRuntimeService 管理插件二进制、运行会话、状态和调用门禁。
type PluginRuntimeService struct {
	repo      *repository.Repository
	pluginSvc *PluginService
	govSvc    *GovernanceService

	mu        sync.Mutex
	processes map[string]*pluginProcess
}

func NewPluginRuntimeService(repo *repository.Repository, pluginSvc *PluginService, govSvc *GovernanceService) *PluginRuntimeService {
	return &PluginRuntimeService{
		repo:      repo,
		pluginSvc: pluginSvc,
		govSvc:    govSvc,
		processes: make(map[string]*pluginProcess),
	}
}

// ResolveBinary 选择当前 OS/ARCH 的插件二进制并校验 sha256。
func (s *PluginRuntimeService) ResolveBinary(manifest pluginapi.Manifest, releaseDir string) (ResolvedPluginBinary, error) {
	if strings.TrimSpace(manifest.ID) == "" {
		return ResolvedPluginBinary{}, errors.New("插件ID不能为空")
	}
	if strings.TrimSpace(releaseDir) == "" {
		return ResolvedPluginBinary{}, errors.New("插件版本目录不能为空")
	}

	var selected pluginapi.BinaryInfo
	for _, binary := range manifest.Package.Binaries {
		platformOK := binary.Platform == runtime.GOOS || binary.Platform == "all"
		archOK := binary.Arch == runtime.GOARCH || binary.Arch == "all"
		if platformOK && archOK && strings.TrimSpace(binary.Path) != "" {
			selected = binary
			break
		}
	}
	if selected.Path == "" && manifest.Backend.EntryExecutable != "" {
		selected = pluginapi.BinaryInfo{
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
			Path:     manifest.Backend.EntryExecutable,
		}
	}
	if selected.Path == "" {
		return ResolvedPluginBinary{}, errors.New("当前平台没有可用插件二进制")
	}

	binaryPath := filepath.Clean(filepath.Join(releaseDir, selected.Path))
	releaseRoot := filepath.Clean(releaseDir)
	if !strings.HasPrefix(binaryPath, releaseRoot) {
		return ResolvedPluginBinary{}, errors.New("插件二进制路径越界")
	}
	hash, err := fileSHA256Hex(binaryPath)
	if err != nil {
		return ResolvedPluginBinary{}, err
	}
	expected := strings.TrimSpace(manifest.Integrity.PackageDigest)
	if selected.Extensions != nil {
		if value, ok := selected.Extensions["sha256"].(string); ok && strings.TrimSpace(value) != "" {
			expected = strings.TrimSpace(value)
		}
	}
	if expected != "" && !strings.EqualFold(expected, hash) {
		return ResolvedPluginBinary{}, errors.New("插件二进制 hash 不匹配")
	}

	return ResolvedPluginBinary{
		PluginID:   manifest.ID,
		Version:    manifest.Version,
		ReleaseDir: releaseRoot,
		Path:       binaryPath,
		SHA256:     hash,
		Platform:   defaultString(selected.Platform, runtime.GOOS),
		Arch:       defaultString(selected.Arch, runtime.GOARCH),
	}, nil
}

// Start 启动插件进程并写入运行会话。
func (s *PluginRuntimeService) Start(ctx context.Context, manifest pluginapi.Manifest, releaseDir string) (*model.PluginRuntimeSession, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("插件运行时服务未初始化")
	}
	binary, err := s.ResolveBinary(manifest, releaseDir)
	if err != nil {
		return nil, err
	}
	instanceID := "pin_" + uuid.NewString()
	cmd := exec.CommandContext(ctx, binary.Path)
	cmd.Dir = releaseDir
	cmd.Env = append(os.Environ(),
		"KEYSWIFT_PLUGIN_ID="+manifest.ID,
		"KEYSWIFT_PLUGIN_VERSION="+manifest.Version,
		"KEYSWIFT_PLUGIN_INSTANCE_ID="+instanceID,
		"KEYSWIFT_PLUGIN_PROTOCOL="+pluginapi.HandshakeProtocol,
	)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	now := time.Now()
	session := &model.PluginRuntimeSession{
		PluginID:   manifest.ID,
		Version:    manifest.Version,
		InstanceID: instanceID,
		PID:        cmd.Process.Pid,
		State:      PluginRuntimeStateStarting,
		StartedAt:  now,
	}
	if err := s.repo.GetDB().Create(session).Error; err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}

	s.mu.Lock()
	s.processes[manifest.ID] = &pluginProcess{cmd: cmd, instanceID: instanceID}
	s.mu.Unlock()
	s.updateRegistryRuntime(manifest.ID, session)
	_ = s.recordState(manifest.ID, "", PluginRuntimeStateStarting, "process.started", "")

	go s.waitProcess(manifest.ID, instanceID, cmd)
	return session, nil
}

// MarkReady 将插件会话切换为 ready，并打开流量门禁。
func (s *PluginRuntimeService) MarkReady(pluginID, instanceID string) error {
	return s.transition(pluginID, instanceID, PluginRuntimeStateReady, "runtime.ready", func(session *model.PluginRuntimeSession) {
		now := time.Now()
		session.ReadyAt = &now
		session.LastHeartbeatAt = &now
	})
}

// Heartbeat 记录插件心跳。
func (s *PluginRuntimeService) Heartbeat(pluginID, instanceID string) error {
	return s.transition(pluginID, instanceID, "", "runtime.heartbeat", func(session *model.PluginRuntimeSession) {
		now := time.Now()
		session.LastHeartbeatAt = &now
	})
}

// Stop 停止插件进程并关闭调用门禁。
func (s *PluginRuntimeService) Stop(pluginID, reason string) error {
	s.mu.Lock()
	process := s.processes[pluginID]
	delete(s.processes, pluginID)
	s.mu.Unlock()
	if process != nil && process.cmd != nil && process.cmd.Process != nil {
		_ = process.cmd.Process.Kill()
	}
	return s.transition(pluginID, "", PluginRuntimeStateStopped, "runtime.stopped", func(session *model.PluginRuntimeSession) {
		now := time.Now()
		session.StoppedAt = &now
		session.FaultReason = reason
	})
}

// Kill 强制结束插件进程，并将会话标记为崩溃。
func (s *PluginRuntimeService) Kill(pluginID, reason string) error {
	s.mu.Lock()
	process := s.processes[pluginID]
	delete(s.processes, pluginID)
	s.mu.Unlock()
	if process != nil && process.cmd != nil && process.cmd.Process != nil {
		_ = process.cmd.Process.Kill()
	}
	return s.transition(pluginID, "", PluginRuntimeStateCrashed, "runtime.killed", func(session *model.PluginRuntimeSession) {
		now := time.Now()
		session.StoppedAt = &now
		session.FaultReason = reason
	})
}

// Restart 停止当前插件进程后按同一 manifest 重新启动。
func (s *PluginRuntimeService) Restart(ctx context.Context, manifest pluginapi.Manifest, releaseDir string) (*model.PluginRuntimeSession, error) {
	_ = s.Stop(manifest.ID, "restart")
	return s.Start(ctx, manifest, releaseDir)
}

func (s *PluginRuntimeService) waitProcess(pluginID, instanceID string, cmd *exec.Cmd) {
	err := cmd.Wait()
	s.mu.Lock()
	current := s.processes[pluginID]
	if current != nil && current.instanceID == instanceID {
		delete(s.processes, pluginID)
	}
	s.mu.Unlock()
	if err != nil {
		_ = s.transition(pluginID, instanceID, PluginRuntimeStateCrashed, "runtime.crashed", func(session *model.PluginRuntimeSession) {
			now := time.Now()
			session.StoppedAt = &now
			session.FaultReason = err.Error()
		})
		if s.govSvc != nil {
			_ = s.govSvc.RecordPluginFault(pluginID, instanceID, "process_exit", err.Error(), "")
		}
	}
}

// Invoke 检查插件 Ready、隔离和超时状态，并记录通过宿主网关发起的插件调用。
func (s *PluginRuntimeService) Invoke(ctx context.Context, req PluginInvocationRequest) (PluginInvocationResult, error) {
	if strings.TrimSpace(req.PluginID) == "" {
		return PluginInvocationResult{}, errors.New("插件ID不能为空")
	}
	session, err := s.latestSession(req.PluginID)
	if err != nil {
		return PluginInvocationResult{}, err
	}
	if session.State != PluginRuntimeStateReady || session.ReadyAt == nil {
		return PluginInvocationResult{Accepted: false, Message: "插件未就绪"}, nil
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	select {
	case <-callCtx.Done():
		return PluginInvocationResult{Accepted: false, Message: "插件调用超时"}, callCtx.Err()
	default:
	}
	if s.govSvc != nil {
		_ = s.govSvc.RecordAudit(AuditInput{
			ActorSubjectID: req.SubjectID,
			Action:         "plugin.invoke",
			ResourceType:   "plugin",
			ResourceID:     req.PluginID,
			RiskLevel:      DefaultPermissionRiskLevel,
			Payload: map[string]any{
				"route_id": req.RouteID,
			},
		})
	}
	return PluginInvocationResult{Accepted: true, Message: "调用已通过宿主门禁"}, nil
}

func (s *PluginRuntimeService) ListSessions(pluginID string) ([]model.PluginRuntimeSession, error) {
	var records []model.PluginRuntimeSession
	query := s.repo.GetDB().Model(&model.PluginRuntimeSession{})
	if strings.TrimSpace(pluginID) != "" {
		query = query.Where("plugin_id = ?", strings.TrimSpace(pluginID))
	}
	err := query.Order("id DESC").Limit(200).Find(&records).Error
	return records, err
}

func (s *PluginRuntimeService) transition(pluginID, instanceID, toState, eventType string, mutate func(*model.PluginRuntimeSession)) error {
	session, err := s.latestSessionForInstance(pluginID, instanceID)
	if err != nil {
		return err
	}
	fromState := session.State
	if toState != "" {
		session.State = toState
	}
	if mutate != nil {
		mutate(session)
	}
	if err := s.repo.GetDB().Save(session).Error; err != nil {
		return err
	}
	s.updateRegistryRuntime(pluginID, session)
	return s.recordState(pluginID, fromState, session.State, eventType, session.FaultReason)
}

func (s *PluginRuntimeService) latestSession(pluginID string) (*model.PluginRuntimeSession, error) {
	return s.latestSessionForInstance(pluginID, "")
}

func (s *PluginRuntimeService) latestSessionForInstance(pluginID, instanceID string) (*model.PluginRuntimeSession, error) {
	var session model.PluginRuntimeSession
	query := s.repo.GetDB().Where("plugin_id = ?", strings.TrimSpace(pluginID))
	if strings.TrimSpace(instanceID) != "" {
		query = query.Where("instance_id = ?", strings.TrimSpace(instanceID))
	}
	if err := query.Order("id DESC").First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("插件运行会话不存在")
		}
		return nil, err
	}
	return &session, nil
}

func (s *PluginRuntimeService) updateRegistryRuntime(pluginID string, session *model.PluginRuntimeSession) {
	if s.pluginSvc == nil || session == nil {
		return
	}
	runtimeInfo, _ := s.pluginSvc.Registry().GetRuntime(pluginID)
	runtimeInfo.PluginID = pluginID
	runtimeInfo.InstanceID = session.InstanceID
	runtimeInfo.PID = session.PID
	runtimeInfo.State = session.State
	runtimeInfo.TrafficEnabled = session.State == PluginRuntimeStateReady && session.ReadyAt != nil
	runtimeInfo.StartedAt = session.StartedAt
	if session.ReadyAt != nil {
		runtimeInfo.ReadyAt = *session.ReadyAt
	}
	if session.LastHeartbeatAt != nil {
		runtimeInfo.LastHeartbeatAt = *session.LastHeartbeatAt
	}
	s.pluginSvc.Registry().SetRuntime(pluginID, runtimeInfo)
}

func (s *PluginRuntimeService) recordState(pluginID, fromState, toState, eventType, reason string) error {
	if s.govSvc == nil {
		return nil
	}
	return s.govSvc.RecordPluginStateEvent(pluginID, fromState, toState, eventType, reason, "")
}

func fileSHA256Hex(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
