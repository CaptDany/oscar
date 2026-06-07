package person

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPersonFullName(t *testing.T) {
	p := &Person{
		FirstName: "John",
		LastName:  "Doe",
	}

	if got := p.FullName(); got != "John Doe" {
		t.Errorf("FullName() = %q, want %q", got, "John Doe")
	}
}

func TestPersonPrimaryEmail(t *testing.T) {
	tests := []struct {
		name     string
		person   *Person
		want     string
	}{
		{"multiple emails", &Person{Email: []string{"primary@test.com", "secondary@test.com"}}, "primary@test.com"},
		{"single email", &Person{Email: []string{"only@test.com"}}, "only@test.com"},
		{"no email", &Person{Email: []string{}}, ""},
		{"nil email", &Person{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.person.PrimaryEmail(); got != tt.want {
				t.Errorf("PrimaryEmail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPersonPrimaryPhone(t *testing.T) {
	tests := []struct {
		name   string
		person *Person
		want   string
	}{
		{"multiple phones", &Person{Phone: []string{"+1-555-0100", "+1-555-0200"}}, "+1-555-0100"},
		{"no phone", &Person{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.person.PrimaryPhone(); got != tt.want {
				t.Errorf("PrimaryPhone() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPersonConstants(t *testing.T) {
	if PersonTypeLead != "lead" {
		t.Errorf("PersonTypeLead = %q", PersonTypeLead)
	}
	if PersonTypeContact != "contact" {
		t.Errorf("PersonTypeContact = %q", PersonTypeContact)
	}
	if PersonTypeCustomer != "customer" {
		t.Errorf("PersonTypeCustomer = %q", PersonTypeCustomer)
	}
}

func TestPersonStruct(t *testing.T) {
	now := time.Now()
	uid := uuid.New()
	tid := uuid.New()
	cid := uuid.New()
	oid := uuid.New()
	source := PersonSource(PersonSourceReferral)
	avatar := "https://example.com/avatar.png"

	p := &Person{
		ID:           uid,
		TenantID:     tid,
		Type:         PersonTypeLead,
		Status:       PersonStatusNew,
		FirstName:    "Alice",
		LastName:     "Smith",
		Email:        []string{"alice@test.com"},
		Phone:        []string{"+1-555-0100"},
		AvatarURL:    &avatar,
		CompanyID:    &cid,
		CompanyName:  strPtr("Acme Corp"),
		OwnerID:      &oid,
		Source:       &source,
		Score:        42,
		Tags:         []string{"vip", "enterprise"},
		ConvertedAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if p.ID != uid {
		t.Error("ID mismatch")
	}
	if p.Type != PersonTypeLead {
		t.Errorf("Type = %q", PersonTypeLead)
	}
	if p.FullName() != "Alice Smith" {
		t.Errorf("FullName = %q", p.FullName())
	}
	if p.PrimaryEmail() != "alice@test.com" {
		t.Errorf("PrimaryEmail = %q", p.PrimaryEmail())
	}
	if len(p.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(p.Tags))
	}
	if p.Score != 42 {
		t.Errorf("Score = %d, want 42", p.Score)
	}
}

func TestPersonStatusConstants(t *testing.T) {
	tests := []struct {
		status PersonStatus
		want   string
	}{
		{PersonStatusNew, "new"},
		{PersonStatusContacted, "contacted"},
		{PersonStatusQualified, "qualified"},
		{PersonStatusUnqualified, "unqualified"},
		{PersonStatusActive, "active"},
		{PersonStatusInactive, "inactive"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("PersonStatus(%s) = %q, want %q", tt.want, string(tt.status), tt.want)
		}
	}
}

func TestPersonSourceConstants(t *testing.T) {
	sources := []PersonSource{
		PersonSourceWebsite, PersonSourceReferral, PersonSourceSocial,
		PersonSourceSocialMedia, PersonSourceEmail, PersonSourceEmailCampaign,
		PersonSourcePhone, PersonSourceColdOutreach, PersonSourceEvent,
		PersonSourcePartner, PersonSourceOther,
	}

	if len(sources) != 11 {
		t.Errorf("expected 11 sources, got %d", len(sources))
	}
}

func TestCreatePersonRequestDefaults(t *testing.T) {
	req := &CreatePersonRequest{
		FirstName: "Bob",
		LastName:  "Jones",
		Type:      PersonTypeLead,
	}

	if req.FirstName != "Bob" {
		t.Errorf("FirstName = %q", req.FirstName)
	}
	if req.Type != PersonTypeLead {
		t.Errorf("Type = %q", req.Type)
	}
	if req.Email != nil {
		t.Errorf("expected nil Email")
	}
}

func TestListPersonsFilter(t *testing.T) {
	oid := uuid.New()
	filter := &ListPersonsFilter{
		Type:      PersonTypeCustomer,
		Status:    PersonStatusActive,
		OwnerID:   &oid,
		Search:    "test",
		Tags:      []string{"vip"},
		Cursor:    "abc123",
		Limit:     50,
	}

	if filter.Limit != 50 {
		t.Errorf("Limit = %d", filter.Limit)
	}
	if filter.Search != "test" {
		t.Errorf("Search = %q", filter.Search)
	}
	if *filter.OwnerID != oid {
		t.Error("OwnerID mismatch")
	}
}

func strPtr(s string) *string {
	return &s
}
