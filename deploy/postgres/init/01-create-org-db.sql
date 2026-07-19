-- PostgreSQL init script: Create vedo_org database for the org service.
-- This script runs automatically when the postgres container starts for the first time.
-- Uses template0 to ensure clean UTF8 encoding.

CREATE DATABASE vedo_org
    WITH ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8'
    TEMPLATE template0;
