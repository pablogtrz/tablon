-- Normalize timestamps to a fixed-width nine-digit RFC3339 form so lexicographic
-- TEXT ordering is chronological. Earlier rows used RFC3339Nano, which trims
-- trailing zeros and can misorder same-second values. Existing rows keep
-- millisecond precision.

UPDATE projects
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z',
    updated_at = strftime('%Y-%m-%dT%H:%M:%f', updated_at) || '000000Z';

UPDATE epics
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z',
    updated_at = strftime('%Y-%m-%dT%H:%M:%f', updated_at) || '000000Z';

UPDATE issues
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z',
    updated_at = strftime('%Y-%m-%dT%H:%M:%f', updated_at) || '000000Z';

UPDATE comments
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z',
    updated_at = strftime('%Y-%m-%dT%H:%M:%f', updated_at) || '000000Z';

UPDATE project_identifiers
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z';

UPDATE idempotency_keys
SET created_at = strftime('%Y-%m-%dT%H:%M:%f', created_at) || '000000Z';
