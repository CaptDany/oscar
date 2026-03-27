package automation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TriggerType string

const (
	TriggerPersonCreated      TriggerType = "person.created"
	TriggerPersonUpdated      TriggerType = "person.updated"
	TriggerPersonConverted    TriggerType = "person.converted"
	TriggerPersonScoreChanged TriggerType = "person.score_changed"
	TriggerPersonAssigned     TriggerType = "person.assigned"
	TriggerDealCreated        TriggerType = "deal.created"
	TriggerDealUpdated        TriggerType = "deal.updated"
	TriggerDealStageChanged   TriggerType = "deal.stage_changed"
	TriggerDealWon            TriggerType = "deal.won"
	TriggerDealLost          TriggerType = "deal.lost"
	TriggerDealCloseDatePassed TriggerType = "deal.close_date_passed"
	TriggerActivityCreated    TriggerType = "activity.created"
	TriggerActivityCompleted  TriggerType = "activity.completed"
	TriggerCompanyCreated     TriggerType = "company.created"
	TriggerCompanyUpdated     TriggerType = "company.updated"
)

type ActionType string

const (
	ActionCreateTask       ActionType = "create_task"
	ActionSendEmail        ActionType = "send_email"
	ActionUpdateField      ActionType = "update_field"
	ActionAddTag           ActionType = "add_tag"
	ActionRemoveTag        ActionType = "remove_tag"
	ActionAssignOwner      ActionType = "assign_owner"
	ActionMoveStage        ActionType = "move_stage"
	ActionConvertPerson    ActionType = "convert_person"
	ActionSendNotification ActionType = "send_notification"
	ActionWebhook          ActionType = "webhook"
	ActionSendSMS          ActionType = "send_sms"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

type ConditionOperator string

const (
	ConditionAnd ConditionOperator = "AND"
	ConditionOr  ConditionOperator = "OR"
)

type Automation struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	Name           string          `json:"name"`
	Description    *string         `json:"description"`
	IsActive       bool            `json:"is_active"`
	TriggerType    TriggerType     `json:"trigger_type"`
	TriggerConfig  json.RawMessage `json:"trigger_config"`
	Conditions     json.RawMessage `json:"conditions"`
	CreatedBy      *uuid.UUID      `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type AutomationAction struct {
	ID            uuid.UUID       `json:"id"`
	AutomationID  uuid.UUID       `json:"automation_id"`
	Position      int             `json:"position"`
	ActionType    ActionType      `json:"action_type"`
	ActionConfig  json.RawMessage `json:"action_config"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type AutomationRun struct {
	ID                 uuid.UUID  `json:"id"`
	AutomationID       uuid.UUID  `json:"automation_id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	TriggerEntityType  *string    `json:"trigger_entity_type"`
	TriggerEntityID    *uuid.UUID `json:"trigger_entity_id"`
	Status             RunStatus  `json:"status"`
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	Error              *string    `json:"error"`
	CreatedAt          time.Time  `json:"created_at"`
}

type AutomationRunAction struct {
	ID          uuid.UUID       `json:"id"`
	RunID       uuid.UUID       `json:"run_id"`
	ActionID    uuid.UUID       `json:"action_id"`
	Status      RunStatus       `json:"status"`
	Result      json.RawMessage `json:"result"`
	ExecutedAt  *time.Time      `json:"executed_at"`
	Error       *string         `json:"error"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Condition struct {
	Operator ConditionOperator `json:"operator"`
	Rules    []ConditionRule  `json:"rules"`
}

type ConditionRule struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type CreateAutomationRequest struct {
	Name          string          `json:"name" validate:"required,min=1,max=255"`
	Description   *string         `json:"description"`
	IsActive      bool            `json:"is_active"`
	TriggerType   TriggerType     `json:"trigger_type" validate:"required"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	Conditions    json.RawMessage `json:"conditions"`
	Actions       []ActionConfig  `json:"actions"`
}

type ActionConfig struct {
	Position     int             `json:"position"`
	ActionType   ActionType      `json:"action_type" validate:"required"`
	ActionConfig json.RawMessage `json:"action_config"`
}

type UpdateAutomationRequest struct {
	Name          *string          `json:"name"`
	Description   *string          `json:"description"`
	IsActive      *bool            `json:"is_active"`
	TriggerType   *TriggerType     `json:"trigger_type"`
	TriggerConfig *json.RawMessage `json:"trigger_config"`
	Conditions    *json.RawMessage `json:"conditions"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateAutomationRequest, createdBy *uuid.UUID) (*Automation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Automation, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Automation, int, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateAutomationRequest) (*Automation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveByTrigger(ctx context.Context, tenantID uuid.UUID, trigger TriggerType) ([]*Automation, error)
}

type ActionRepository interface {
	Create(ctx context.Context, automationID uuid.UUID, actions []ActionConfig) ([]AutomationAction, error)
	ListByAutomation(ctx context.Context, automationID uuid.UUID) ([]AutomationAction, error)
	Update(ctx context.Context, id uuid.UUID, actionType *ActionType, config *json.RawMessage) (*AutomationAction, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAutomation(ctx context.Context, automationID uuid.UUID) error
}

type RunRepository interface {
	Create(ctx context.Context, automationID, tenantID uuid.UUID, entityType *string, entityID *uuid.UUID) (*AutomationRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AutomationRun, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*AutomationRun, int, error)
	ListByAutomation(ctx context.Context, automationID uuid.UUID, limit, offset int) ([]*AutomationRun, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status RunStatus, startedAt, completedAt *time.Time, err *string) (*AutomationRun, error)
	CreateRunAction(ctx context.Context, runID, actionID uuid.UUID) (*AutomationRunAction, error)
	UpdateRunAction(ctx context.Context, id uuid.UUID, status RunStatus, result json.RawMessage, err *string) error
}
