package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kamilrybacki/edictflow/server/domain"
)

// GraphTeamService defines the team operations needed for graph data
type GraphTeamService interface {
	List() ([]domain.Team, error)
}

// GraphUserService defines the user operations needed for graph data
type GraphUserService interface {
	List(teamID string, activeOnly bool) ([]domain.User, error)
	CountByTeam(teamID string) (int, error)
}

// GraphRuleService defines the rule operations needed for graph data
type GraphRuleService interface {
	ListAll() ([]domain.Rule, error)
}

// GraphTeam represents a team in the graph response
type GraphTeam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

// GraphUser represents a user in the graph response
type GraphUser struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	TeamID *string `json:"teamId"`
}

// GraphRule represents a rule in the graph response
type GraphRule struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Status          domain.RuleStatus `json:"status"`
	EnforcementMode string            `json:"enforcementMode"`
	TeamID          *string           `json:"teamId"`
	TargetTeams     []string          `json:"targetTeams"`
	TargetUsers     []string          `json:"targetUsers"`
}

// GraphResponse is the complete graph data response
type GraphResponse struct {
	Teams []GraphTeam `json:"teams"`
	Users []GraphUser `json:"users"`
	Rules []GraphRule `json:"rules"`
}

// GraphHandler handles graph-related API requests
type GraphHandler struct {
	teamService GraphTeamService
	userService GraphUserService
	ruleService GraphRuleService
}

// NewGraphHandler creates a new GraphHandler
func NewGraphHandler(
	teamService GraphTeamService,
	userService GraphUserService,
	ruleService GraphRuleService,
) *GraphHandler {
	return &GraphHandler{
		teamService: teamService,
		userService: userService,
		ruleService: ruleService,
	}
}

// Get returns all graph data in a single response
func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Fetch all teams
	teams, err := h.teamService.List()
	if err != nil {
		http.Error(w, "Failed to fetch teams", http.StatusInternalServerError)
		return
	}

	// Fetch all users
	users, err := h.userService.List("", true)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Fetch all rules
	rules, err := h.ruleService.ListAll()
	if err != nil {
		http.Error(w, "Failed to fetch rules", http.StatusInternalServerError)
		return
	}

	// Build response
	response := GraphResponse{
		Teams: make([]GraphTeam, 0, len(teams)),
		Users: make([]GraphUser, 0, len(users)),
		Rules: make([]GraphRule, 0, len(rules)),
	}

	// Transform teams with member counts
	for _, team := range teams {
		count, _ := h.userService.CountByTeam(team.ID)
		response.Teams = append(response.Teams, GraphTeam{
			ID:          team.ID,
			Name:        team.Name,
			MemberCount: count,
		})
	}

	// Transform users
	for _, user := range users {
		response.Users = append(response.Users, GraphUser{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			TeamID: user.TeamID,
		})
	}

	// Transform rules
	for _, rule := range rules {
		targetTeams := rule.TargetTeams
		if targetTeams == nil {
			targetTeams = []string{}
		}
		targetUsers := rule.TargetUsers
		if targetUsers == nil {
			targetUsers = []string{}
		}

		response.Rules = append(response.Rules, GraphRule{
			ID:              rule.ID,
			Name:            rule.Name,
			Status:          rule.Status,
			EnforcementMode: string(rule.EnforcementMode),
			TeamID:          rule.TeamID,
			TargetTeams:     targetTeams,
			TargetUsers:     targetUsers,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
