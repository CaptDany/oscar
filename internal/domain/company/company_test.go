package company

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCompanyStruct(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	tid := uuid.New()
	oid := uuid.New()
	pid := uuid.New()
	rev := 1_000_000.50
	domain := "acme.com"
	industry := "Technology"
	size := CompanySize(CompanySizeMedium)

	c := &Company{
		ID:              id,
		TenantID:        tid,
		Name:            "Acme Corp",
		Domain:          &domain,
		Industry:        &industry,
		Size:            &size,
		AnnualRevenue:   &rev,
		Website:         strPtr("https://acme.com"),
		Address:         map[string]string{"city": "San Francisco"},
		OwnerID:         &oid,
		ParentCompanyID: &pid,
		Tags:            []string{"enterprise", "tech"},
		CustomFields:    map[string]interface{}{"tier": "platinum"},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if c.ID != id {
		t.Error("ID mismatch")
	}
	if c.Name != "Acme Corp" {
		t.Errorf("Name = %q", c.Name)
	}
	if *c.Domain != "acme.com" {
		t.Errorf("Domain = %q", *c.Domain)
	}
	if *c.AnnualRevenue != rev {
		t.Errorf("AnnualRevenue = %f", *c.AnnualRevenue)
	}
	if len(c.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(c.Tags))
	}
}

func TestCompanySizeValues(t *testing.T) {
	tests := []struct {
		size CompanySize
		want string
	}{
		{CompanySizeStartup, "startup"},
		{CompanySizeSmall, "small"},
		{CompanySizeMedium, "medium"},
		{CompanySizeLarge, "large"},
		{CompanySizeEnterprise, "enterprise"},
	}

	for _, tt := range tests {
		if string(tt.size) != tt.want {
			t.Errorf("CompanySize(%s) = %q, want %q", tt.want, string(tt.size), tt.want)
		}
	}
}

func TestCreateCompanyRequest(t *testing.T) {
	req := &CreateCompanyRequest{
		Name: "Test Corp",
		Tags: []string{"test"},
	}

	if req.Name != "Test Corp" {
		t.Errorf("Name = %q", req.Name)
	}
	if len(req.Tags) != 1 {
		t.Errorf("expected 1 tag")
	}
}

func TestUpdateCompanyRequest(t *testing.T) {
	newName := "Updated Corp"
	newIndustry := "Healthcare"
	req := &UpdateCompanyRequest{
		Name:     &newName,
		Industry: &newIndustry,
	}

	if *req.Name != "Updated Corp" {
		t.Errorf("Name = %q", *req.Name)
	}
	if *req.Industry != "Healthcare" {
		t.Errorf("Industry = %q", *req.Industry)
	}
}

func TestListCompaniesFilter(t *testing.T) {
	now := time.Now()
	oid := uuid.New()
	cid := uuid.New()

	filter := &ListCompaniesFilter{
		OwnerID:      &oid,
		Search:       "acme",
		Cursor:       "cursor",
		CursorAfter:  &now,
		CursorID:     &cid,
		Limit:        25,
		IncludeTotal: true,
	}

	if filter.Limit != 25 {
		t.Errorf("Limit = %d", filter.Limit)
	}
	if !filter.IncludeTotal {
		t.Error("IncludeTotal should be true")
	}
	if *filter.OwnerID != oid {
		t.Error("OwnerID mismatch")
	}
}

func TestAddressStruct(t *testing.T) {
	addr := Address{
		Street:     "123 Main St",
		City:       "San Francisco",
		State:      "CA",
		PostalCode: "94105",
		Country:    "US",
	}

	if addr.City != "San Francisco" {
		t.Errorf("City = %q", addr.City)
	}
	if addr.Country != "US" {
		t.Errorf("Country = %q", addr.Country)
	}
}

func TestCompanyNilableFields(t *testing.T) {
	c := &Company{
		ID:   uuid.New(),
		Name: "Minimal",
	}

	if c.Domain != nil {
		t.Error("expected nil Domain")
	}
	if c.Industry != nil {
		t.Error("expected nil Industry")
	}
	if c.AnnualRevenue != nil {
		t.Error("expected nil AnnualRevenue")
	}
	if c.OwnerID != nil {
		t.Error("expected nil OwnerID")
	}
}

func strPtr(s string) *string {
	return &s
}
