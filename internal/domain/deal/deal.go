package deal

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Deal struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Title             string     `json:"title"`
	Value             float64    `json:"value"`
	Currency          string     `json:"currency"`
	StageID           *uuid.UUID `json:"stage_id"`
	PipelineID        *uuid.UUID `json:"pipeline_id"`
	PersonID          *uuid.UUID `json:"person_id"`
	CompanyID         *uuid.UUID `json:"company_id"`
	OwnerID           *uuid.UUID `json:"owner_id"`
	ExpectedCloseDate *time.Time `json:"expected_close_date"`
	ClosedAt          *time.Time `json:"closed_at"`
	WonReason         *string    `json:"won_reason"`
	LostReason        *string    `json:"lost_reason"`
	Probability       int        `json:"probability"`
	Tags              []string    `json:"tags"`
	CustomFields      interface{} `json:"custom_fields"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"-"`
}

type PipelineStage struct {
	ID          uuid.UUID `json:"id"`
	PipelineID  uuid.UUID `json:"pipeline_id"`
	Name        string    `json:"name"`
	Position    int       `json:"position"`
	Probability int       `json:"probability"`
	StageType   string    `json:"stage_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Pipeline struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	Currency  string          `json:"currency"`
	Stages    []PipelineStage `json:"stages,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type PipelineStats struct {
	StageID     uuid.UUID `json:"stage_id"`
	StageName   string    `json:"stage_name"`
	Probability int       `json:"probability"`
	DealCount   int       `json:"deal_count"`
	TotalValue  float64   `json:"total_value"`
}

type DealWithStage struct {
	Deal
	StageName     string `json:"stage_name"`
	StagePosition int    `json:"stage_position"`
	StageType     string `json:"stage_type"`
}

type CreateDealRequest struct {
	Title             string        `json:"title" validate:"required,min=1,max=255"`
	Value             float64       `json:"value"`
	Currency          string        `json:"currency"`
	StageID           *uuid.UUID    `json:"stage_id"`
	PipelineID        *uuid.UUID    `json:"pipeline_id"`
	PersonID          *uuid.UUID    `json:"person_id"`
	CompanyID         *uuid.UUID    `json:"company_id"`
	OwnerID           *uuid.UUID    `json:"owner_id"`
	ExpectedCloseDate *time.Time    `json:"expected_close_date"`
	Tags              []string      `json:"tags"`
	CustomFields      interface{}   `json:"custom_fields"`
}

type UpdateDealRequest struct {
	Title             *string       `json:"title"`
	Value             *float64      `json:"value"`
	Currency          *string       `json:"currency"`
	StageID           *uuid.UUID    `json:"stage_id"`
	PipelineID        *uuid.UUID    `json:"pipeline_id"`
	PersonID          *uuid.UUID    `json:"person_id"`
	CompanyID         *uuid.UUID    `json:"company_id"`
	OwnerID           *uuid.UUID    `json:"owner_id"`
	ExpectedCloseDate *time.Time    `json:"expected_close_date"`
	Probability       *int          `json:"probability"`
	Tags              []string      `json:"tags"`
	CustomFields      interface{}   `json:"custom_fields"`
	ClosedAt          *time.Time    `json:"closed_at"`
	WonReason         *string       `json:"won_reason"`
	LostReason        *string       `json:"lost_reason"`
}

type MoveDealRequest struct {
	StageID     uuid.UUID `json:"stage_id" validate:"required"`
	Probability int       `json:"probability"`
}

type ListDealsFilter struct {
	StageID    *uuid.UUID
	PipelineID *uuid.UUID
	OwnerID    *uuid.UUID
	PersonID   *uuid.UUID
	Search     string
	Cursor     string
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateDealRequest) (*Deal, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Deal, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateDealRequest) (*Deal, error)
	SoftDelete(ctx context.Context, id uuid.UUID) (*Deal, error)
	MoveToStage(ctx context.Context, id uuid.UUID, stageID, pipelineID uuid.UUID, probability int, closedAt *time.Time) (*Deal, error)
	CloseAsWon(ctx context.Context, id uuid.UUID, stageID uuid.UUID, reason string) (*Deal, error)
	CloseAsLost(ctx context.Context, id uuid.UUID, stageID uuid.UUID, reason string) (*Deal, error)
	List(ctx context.Context, tenantID uuid.UUID, filter *ListDealsFilter) ([]*Deal, string, int, error)
	ListByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) ([]*DealWithStage, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*Deal, error)
	Count(ctx context.Context, tenantID uuid.UUID, filter *ListDealsFilter) (int, error)
	GetByCloseDate(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*Deal, error)
	GetPipelineStats(ctx context.Context, pipelineID uuid.UUID) ([]PipelineStats, error)
}

type PipelineRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, name string, isDefault bool, currency string) (*Pipeline, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Pipeline, error)
	GetDefault(ctx context.Context, tenantID uuid.UUID) (*Pipeline, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Pipeline, error)
	Update(ctx context.Context, id uuid.UUID, name *string, isDefault *bool, currency *string) (*Pipeline, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefault(ctx context.Context, tenantID, pipelineID uuid.UUID) error
	CreateStage(ctx context.Context, pipelineID uuid.UUID, name string, position, probability int, stageType string) (*PipelineStage, error)
	GetStageByID(ctx context.Context, id uuid.UUID) (*PipelineStage, error)
	ListStages(ctx context.Context, pipelineID uuid.UUID) ([]PipelineStage, error)
	UpdateStage(ctx context.Context, id uuid.UUID, name *string, probability *int, stageType *string) (*PipelineStage, error)
	ReorderStages(ctx context.Context, pipelineID uuid.UUID, stageIDs []uuid.UUID) error
	DeleteStage(ctx context.Context, id uuid.UUID) error
}
