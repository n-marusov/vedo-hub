-- Migration 006: Add MVP role vocabulary to memberships CHECK constraint
-- Adds Guest, Reporter, Developer while keeping legacy roles for backward compatibility.

-- PostgreSQL requires dropping and recreating the constraint.
ALTER TABLE memberships DROP CONSTRAINT IF EXISTS memberships_role_check;

ALTER TABLE memberships ADD CONSTRAINT memberships_role_check
    CHECK (role IN (
        -- MVP roles (user-facing)
        'Guest', 'Reporter', 'Developer', 'Maintainer', 'Owner',
        -- Legacy roles (backward compatibility aliases)
        'Viewer', 'Editor',
        -- Enterprise/ops roles (outside MVP happy path)
        'SupportEngineer', 'SRE', 'SecurityLead', 'ProductOwner'
    ));
