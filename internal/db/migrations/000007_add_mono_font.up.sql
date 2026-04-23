-- Add mono_font column to tenant_branding for mono font selection
ALTER TABLE tenant_branding ADD COLUMN IF NOT EXISTS mono_font VARCHAR(255) DEFAULT 'Geist Mono, monospace';