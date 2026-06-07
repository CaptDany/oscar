package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserFullName(t *testing.T) {
	u := &User{
		FirstName: "John",
		LastName:  "Doe",
	}

	if got := u.FullName(); got != "John Doe" {
		t.Errorf("FullName() = %q, want %q", got, "John Doe")
	}
}

func TestUserHasPermission(t *testing.T) {
	resource := "deals"
	viewPerm := Permission{View: ScopeAll, Create: ScopeOwn, Edit: ScopeOwn, Delete: ScopeNone}
	editPerm := Permission{View: ScopeNone, Create: ScopeNone, Edit: ScopeAll, Delete: ScopeNone}

	ownerRole := Role{
		ID:          uuid.New(),
		Name:        "Owner",
		IsSystem:    true,
		Permissions: map[string]Permission{resource: viewPerm},
	}

	editorRole := Role{
		ID:          uuid.New(),
		Name:        "Editor",
		IsSystem:    false,
		Permissions: map[string]Permission{resource: editPerm},
	}

	noPermRole := Role{
		ID:          uuid.New(),
		Name:        "Read Only",
		IsSystem:    true,
		Permissions: map[string]Permission{},
	}

	u := &User{
		ID:        uuid.New(),
		FirstName: "Jane",
		LastName:  "Smith",
	}

	tests := []struct {
		name     string
		roles    []Role
		resource string
		action   string
		want     bool
	}{
		{"owner can view", []Role{ownerRole}, resource, "view", true},
		{"editor can edit", []Role{editorRole}, resource, "edit", true},
		{"no perm cannot view", []Role{noPermRole}, resource, "view", false},
		{"no perm cannot delete", []Role{noPermRole}, resource, "delete", false},
		{"owner cannot delete", []Role{ownerRole}, resource, "delete", false},
		{"empty roles", []Role{}, resource, "view", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.HasPermission(tt.resource, tt.action, tt.roles)
			if got != tt.want {
				t.Errorf("HasPermission(%q, %q) = %v, want %v", tt.resource, tt.action, got, tt.want)
			}
		})
	}
}

func TestUserHasPermissionExport(t *testing.T) {
	perm := Permission{View: ScopeOwn, Export: ScopeAll}
	role := Role{
		Name:        "Exporter",
		Permissions: map[string]Permission{"reports": perm},
	}

	u := &User{FirstName: "A", LastName: "B"}

	if !u.HasPermission("reports", "export", []Role{role}) {
		t.Error("should have export permission")
	}
	if u.HasPermission("reports", "view", []Role{{Name: "No Perm", Permissions: map[string]Permission{}}}) {
		t.Error("should not have view permission")
	}
}

func TestUserFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	tid := uuid.New()

	u := &User{
		ID:              id,
		TenantID:        tid,
		Email:           "test@example.com",
		PasswordHash:    "hash",
		FirstName:       "Alice",
		LastName:        "Wonderland",
		Timezone:        "UTC",
		Locale:          "en",
		IsActive:        true,
		EmailVerifiedAt: &now,
		LastLoginAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if u.ID != id {
		t.Errorf("ID mismatch")
	}
	if u.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "test@example.com")
	}
	if !u.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestCreateUserRequestValidation(t *testing.T) {
	req := &CreateUserRequest{
		Email:     "user@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Timezone:  "America/New_York",
		Locale:    "en",
	}

	if req.Email != "user@example.com" {
		t.Errorf("Email = %q", req.Email)
	}
	if len(req.Password) < 8 {
		t.Error("Password too short")
	}
	if req.FirstName == "" {
		t.Error("FirstName is required")
	}
	if req.LastName == "" {
		t.Error("LastName is required")
	}
}

func TestRoleStruct(t *testing.T) {
	id := uuid.New()
	desc := "Administrator role"
	role := Role{
		ID:          id,
		Name:        "Admin",
		Description: &desc,
		IsSystem:    true,
		Permissions: map[string]Permission{"all": {View: ScopeAll, Create: ScopeAll, Edit: ScopeAll, Delete: ScopeAll}},
	}

	if role.ID != id {
		t.Error("ID mismatch")
	}
	if role.Name != "Admin" {
		t.Errorf("Name = %q", role.Name)
	}
	if !role.IsSystem {
		t.Error("IsSystem should be true")
	}
	if role.Permissions["all"].View != ScopeAll {
		t.Errorf("expected ScopeAll, got %q", role.Permissions["all"].View)
	}
}

func TestUserResponseConversion(t *testing.T) {
	u := &User{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Email:     "user@test.com",
		FirstName: "First",
		LastName:  "Last",
		IsActive:  true,
	}

	resp := UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		IsActive:  u.IsActive,
	}

	if resp.Email != u.Email {
		t.Errorf("response email mismatch")
	}
	if resp.FirstName+" "+resp.LastName != "First Last" {
		t.Errorf("unexpected name: %s %s", resp.FirstName, resp.LastName)
	}
}
