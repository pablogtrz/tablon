-- Optional custom project gradient. Empty colours mean the curated tile is used.
ALTER TABLE projects ADD COLUMN logo_gradient_from TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN logo_gradient_to TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN logo_gradient_angle INTEGER NOT NULL DEFAULT 135
    CHECK (logo_gradient_angle BETWEEN 0 AND 360);
