-- name: CreateTeam :one
INSERT INTO teams (tenant_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1;

-- name: ListTeams :many
SELECT * FROM teams WHERE tenant_id = $1 ORDER BY name ASC;

-- name: UpdateTeam :one
UPDATE teams
SET name = COALESCE($2, name), description = COALESCE($3, description)
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :one
DELETE FROM teams WHERE id = $1 RETURNING *;

-- name: AddTeamMember :one
INSERT INTO team_members (team_id, user_id, is_lead)
VALUES ($1, $2, $3)
ON CONFLICT (team_id, user_id) DO UPDATE SET is_lead = $3
RETURNING *;

-- name: RemoveTeamMember :exec
DELETE FROM team_members WHERE team_id = $1 AND user_id = $2;

-- name: ListTeamMembers :many
SELECT tm.*, u.email, u.first_name, u.last_name, u.avatar_url
FROM team_members tm
JOIN users u ON tm.user_id = u.id
WHERE tm.team_id = $1;

-- name: ListUserTeams :many
SELECT t.*, tm.is_lead
FROM teams t
JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = $1;

-- name: SetTeamLead :exec
UPDATE team_members SET is_lead = false WHERE team_id = $1;
UPDATE team_members SET is_lead = true WHERE team_id = $1 AND user_id = $2;
