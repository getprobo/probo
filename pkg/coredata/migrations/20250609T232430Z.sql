ALTER TABLE organizations
    ADD COLUMN founding_year INTEGER,
    ADD COLUMN company_type TEXT,
    ADD COLUMN premarket_fit BOOLEAN,
    ADD COLUMN uses_cloud_providers BOOLEAN,
    ADD COLUMN ai_focused BOOLEAN,
    ADD COLUMN uses_ai_generated_code BOOLEAN,
    ADD COLUMN vc_backed BOOLEAN,
    ADD COLUMN has_raised_money BOOLEAN,
    ADD COLUMN has_enterprise_accounts BOOLEAN;
