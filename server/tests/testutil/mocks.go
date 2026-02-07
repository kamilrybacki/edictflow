package testutil

import (
	"context"
	"errors"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/services/approvals"
	"github.com/kamilrybacki/edictflow/server/services/auth"
)

// teamIDMatches checks if a *string TeamID matches a string value
func teamIDMatches(teamIDPtr *string, teamID string) bool {
	if teamIDPtr == nil {
		return teamID == ""
	}
	return *teamIDPtr == teamID
}

// Common errors for testing
var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
	ErrDatabaseError = errors.New("database error")
	ErrConflict      = errors.New("conflict")
	ErrTimeout       = errors.New("timeout")
)

// MockAuthService implements auth service interface for testing
type MockAuthService struct {
	RegisterFunc func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error)
	LoginFunc    func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error)
}

func (m *MockAuthService) Register(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, req)
	}
	user := domain.User{
		ID:    "test-user-id",
		Email: req.Email,
		Name:  req.Name,
	}
	return "test-token", user, nil
}

func (m *MockAuthService) Login(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, req)
	}
	user := domain.User{
		ID:    "test-user-id",
		Email: req.Email,
		Name:  "Test User",
	}
	return "test-token", user, nil
}

// MockUserServiceForAuth implements user service interface for auth handler
type MockUserServiceForAuth struct {
	Users              map[string]domain.User
	GetByIDFunc        func(ctx context.Context, id string) (domain.User, error)
	UpdateFunc         func(ctx context.Context, user domain.User) error
	UpdatePasswordFunc func(ctx context.Context, userID, oldPassword, newPassword string) error
}

func NewMockUserServiceForAuth() *MockUserServiceForAuth {
	return &MockUserServiceForAuth{Users: make(map[string]domain.User)}
}

func (m *MockUserServiceForAuth) GetByID(ctx context.Context, id string) (domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if user, ok := m.Users[id]; ok {
		return user, nil
	}
	return domain.User{}, ErrNotFound
}

func (m *MockUserServiceForAuth) Update(ctx context.Context, user domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserServiceForAuth) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, oldPassword, newPassword)
	}
	return nil
}

// MockUsersService implements users service interface for testing
type MockUsersService struct {
	Users                         map[string]domain.User
	ListFunc                      func(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error)
	GetByIDFunc                   func(ctx context.Context, id string) (domain.User, error)
	UpdateFunc                    func(ctx context.Context, user domain.User) error
	DeactivateFunc                func(ctx context.Context, id string) error
	GetWithRolesAndPermissionsFunc func(ctx context.Context, id string) (domain.User, error)
}

func NewMockUsersService() *MockUsersService {
	return &MockUsersService{Users: make(map[string]domain.User)}
}

func (m *MockUsersService) GetByID(ctx context.Context, id string) (domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if user, ok := m.Users[id]; ok {
		return user, nil
	}
	return domain.User{}, ErrNotFound
}

func (m *MockUsersService) List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, teamID, activeOnly)
	}
	var result []domain.User
	for _, u := range m.Users {
		if activeOnly && !u.IsActive {
			continue
		}
		result = append(result, u)
	}
	return result, nil
}

func (m *MockUsersService) Update(ctx context.Context, user domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	if _, ok := m.Users[user.ID]; !ok {
		return ErrNotFound
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockUsersService) Deactivate(ctx context.Context, id string) error {
	if m.DeactivateFunc != nil {
		return m.DeactivateFunc(ctx, id)
	}
	if user, ok := m.Users[id]; ok {
		user.IsActive = false
		m.Users[id] = user
		return nil
	}
	return ErrNotFound
}

func (m *MockUsersService) GetWithRolesAndPermissions(ctx context.Context, id string) (domain.User, error) {
	if m.GetWithRolesAndPermissionsFunc != nil {
		return m.GetWithRolesAndPermissionsFunc(ctx, id)
	}
	return m.GetByID(ctx, id)
}

func (m *MockUsersService) LeaveTeam(ctx context.Context, userID string) error {
	if user, ok := m.Users[userID]; ok {
		user.TeamID = nil
		m.Users[userID] = user
		return nil
	}
	return ErrNotFound
}

// MockRolesService implements roles service interface for testing
type MockRolesService struct {
	Roles           map[string]domain.Role
	Permissions     map[string][]domain.Permission
	UserRoles       map[string][]string
	CreateFunc      func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error)
	GetByIDFunc     func(ctx context.Context, id string) (domain.Role, error)
	ListFunc        func(ctx context.Context, teamID *string) ([]domain.Role, error)
	UpdateFunc      func(ctx context.Context, role domain.Role) error
	DeleteFunc      func(ctx context.Context, id string) error
}

func NewMockRolesService() *MockRolesService {
	return &MockRolesService{
		Roles:       make(map[string]domain.Role),
		Permissions: make(map[string][]domain.Permission),
		UserRoles:   make(map[string][]string),
	}
}

func (m *MockRolesService) Create(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, name, description, hierarchyLevel, parentRoleID, teamID)
	}
	role := domain.NewRole(name, description, hierarchyLevel, parentRoleID, teamID)
	m.Roles[role.ID] = role
	return role, nil
}

func (m *MockRolesService) GetByID(ctx context.Context, id string) (domain.Role, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if role, ok := m.Roles[id]; ok {
		return role, nil
	}
	return domain.Role{}, ErrNotFound
}

func (m *MockRolesService) List(ctx context.Context, teamID *string) ([]domain.Role, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, teamID)
	}
	var result []domain.Role
	for _, r := range m.Roles {
		result = append(result, r)
	}
	return result, nil
}

func (m *MockRolesService) Update(ctx context.Context, role domain.Role) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, role)
	}
	if _, ok := m.Roles[role.ID]; !ok {
		return ErrNotFound
	}
	m.Roles[role.ID] = role
	return nil
}

func (m *MockRolesService) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	if _, ok := m.Roles[id]; !ok {
		return ErrNotFound
	}
	delete(m.Roles, id)
	return nil
}

func (m *MockRolesService) GetRoleWithPermissions(ctx context.Context, id string) (domain.Role, error) {
	role, ok := m.Roles[id]
	if !ok {
		return domain.Role{}, ErrNotFound
	}
	role.Permissions = m.Permissions[id]
	return role, nil
}

func (m *MockRolesService) AddPermission(ctx context.Context, roleID, permissionID string) error {
	if _, ok := m.Roles[roleID]; !ok {
		return errors.New("role not found")
	}
	if permissionID == "" {
		return errors.New("permission id required")
	}
	m.Permissions[roleID] = append(m.Permissions[roleID], domain.Permission{ID: permissionID, Code: permissionID})
	return nil
}

func (m *MockRolesService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	perms := m.Permissions[roleID]
	for i, p := range perms {
		if p.ID == permissionID {
			m.Permissions[roleID] = append(perms[:i], perms[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockRolesService) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	if _, ok := m.Roles[roleID]; !ok {
		return errors.New("role not found")
	}
	if userID == "" {
		return errors.New("user id required")
	}
	m.UserRoles[userID] = append(m.UserRoles[userID], roleID)
	return nil
}

func (m *MockRolesService) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	roles := m.UserRoles[userID]
	for i, r := range roles {
		if r == roleID {
			m.UserRoles[userID] = append(roles[:i], roles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockRolesService) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return []domain.Permission{
		{ID: "perm-1", Code: "rules:read", Description: "Read rules"},
		{ID: "perm-2", Code: "rules:write", Description: "Write rules"},
	}, nil
}

// MockApprovalsService implements approvals service interface for testing
type MockApprovalsService struct {
	Rules           map[string]domain.Rule
	ApprovalRecords map[string][]domain.RuleApproval
	SubmitFunc      func(ctx context.Context, ruleID string) error
	ApproveFunc     func(ctx context.Context, ruleID, userID, comment string) error
	RejectFunc      func(ctx context.Context, ruleID, userID, comment string) error
}

func NewMockApprovalsService() *MockApprovalsService {
	return &MockApprovalsService{
		Rules:           make(map[string]domain.Rule),
		ApprovalRecords: make(map[string][]domain.RuleApproval),
	}
}

func (m *MockApprovalsService) SubmitRule(ctx context.Context, ruleID string) error {
	if m.SubmitFunc != nil {
		return m.SubmitFunc(ctx, ruleID)
	}
	rule, ok := m.Rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusDraft {
		return approvals.ErrCannotSubmit
	}
	rule.Status = domain.RuleStatusPending
	m.Rules[ruleID] = rule
	return nil
}

func (m *MockApprovalsService) ApproveRule(ctx context.Context, ruleID, userID, comment string) error {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, ruleID, userID, comment)
	}
	rule, ok := m.Rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusPending {
		return approvals.ErrNotPending
	}
	m.ApprovalRecords[ruleID] = append(m.ApprovalRecords[ruleID], domain.RuleApproval{
		ID:        "approval-1",
		RuleID:    ruleID,
		UserID:    userID,
		Decision:  domain.ApprovalDecisionApproved,
		Comment:   comment,
		CreatedAt: time.Now(),
	})
	return nil
}

func (m *MockApprovalsService) RejectRule(ctx context.Context, ruleID, userID, comment string) error {
	if m.RejectFunc != nil {
		return m.RejectFunc(ctx, ruleID, userID, comment)
	}
	rule, ok := m.Rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusPending {
		return approvals.ErrNotPending
	}
	rule.Status = domain.RuleStatusRejected
	m.Rules[ruleID] = rule
	return nil
}

func (m *MockApprovalsService) GetApprovalStatus(ctx context.Context, ruleID string) (approvals.ApprovalStatus, error) {
	if _, ok := m.Rules[ruleID]; !ok {
		return approvals.ApprovalStatus{}, approvals.ErrRuleNotFound
	}
	return approvals.ApprovalStatus{
		RuleID:        ruleID,
		Status:        m.Rules[ruleID].Status,
		RequiredCount: 2,
		CurrentCount:  len(m.ApprovalRecords[ruleID]),
		Approvals:     m.ApprovalRecords[ruleID],
	}, nil
}

func (m *MockApprovalsService) GetPendingRules(ctx context.Context, teamID string) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.Rules {
		if rule.Status == domain.RuleStatusPending && teamIDMatches(rule.TeamID, teamID) {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (m *MockApprovalsService) GetPendingRulesByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.Rules {
		if rule.Status == domain.RuleStatusPending && rule.TargetLayer == scope {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (m *MockApprovalsService) ResetRule(ctx context.Context, ruleID string) error {
	rule, ok := m.Rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	rule.Status = domain.RuleStatusDraft
	m.Rules[ruleID] = rule
	delete(m.ApprovalRecords, ruleID)
	return nil
}

// MockRuleService implements rule service interface for testing
type MockRuleService struct {
	Rules          map[string]domain.Rule
	CreateFunc     func(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error)
	GetByIDFunc    func(ctx context.Context, id string) (domain.Rule, error)
	ListByTeamFunc func(ctx context.Context, teamID string) ([]domain.Rule, error)
	UpdateFunc     func(ctx context.Context, rule domain.Rule) error
	DeleteFunc     func(ctx context.Context, id string) error
}

func NewMockRuleService() *MockRuleService {
	return &MockRuleService{Rules: make(map[string]domain.Rule)}
}

func (m *MockRuleService) Create(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	var triggers []domain.Trigger
	for _, t := range req.Triggers {
		triggers = append(triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}
	rule := domain.NewRule(req.Name, domain.TargetLayer(req.TargetLayer), req.Content, triggers, req.TeamID)
	m.Rules[rule.ID] = rule
	return rule, nil
}

func (m *MockRuleService) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	rule, ok := m.Rules[id]
	if !ok {
		return domain.Rule{}, handlers.ErrNotFound
	}
	return rule, nil
}

func (m *MockRuleService) ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	if m.ListByTeamFunc != nil {
		return m.ListByTeamFunc(ctx, teamID)
	}
	var result []domain.Rule
	for _, r := range m.Rules {
		if teamIDMatches(r.TeamID, teamID) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRuleService) ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.Rules {
		if teamIDMatches(r.TeamID, teamID) && r.Status == status {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRuleService) Update(ctx context.Context, rule domain.Rule) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, rule)
	}
	m.Rules[rule.ID] = rule
	return nil
}

func (m *MockRuleService) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	delete(m.Rules, id)
	return nil
}

func (m *MockRuleService) ListByTargetLayer(ctx context.Context, targetLayer domain.TargetLayer) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.Rules {
		if r.TargetLayer == targetLayer {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRuleService) GetMergedContent(ctx context.Context, targetLayer domain.TargetLayer) (string, error) {
	rules, _ := m.ListByTargetLayer(ctx, targetLayer)
	if len(rules) == 0 {
		return "", nil
	}
	// Simple mock: just return rule contents concatenated
	var content string
	for _, r := range rules {
		content += r.Content + "\n"
	}
	return content, nil
}

func (m *MockRuleService) ListGlobal(ctx context.Context) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.Rules {
		if r.IsGlobal() {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRuleService) CreateGlobal(ctx context.Context, name, content string, description *string, force bool) (domain.Rule, error) {
	rule := domain.NewGlobalRule(name, content, force)
	rule.Description = description
	m.Rules[rule.ID] = rule
	return rule, nil
}
