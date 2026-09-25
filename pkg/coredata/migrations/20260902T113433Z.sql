-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

ALTER TABLE iam_membership_profiles
    DROP CONSTRAINT iam_membership_profiles_organization_id_fkey;
ALTER TABLE iam_membership_profiles
    ADD CONSTRAINT iam_membership_profiles_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE audit_log_entries
    DROP CONSTRAINT audit_log_entries_organization_id_fkey;
ALTER TABLE audit_log_entries
    ADD CONSTRAINT audit_log_entries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE cookie_banner_translations
    DROP CONSTRAINT cookie_banner_translations_organization_id_fkey;
ALTER TABLE cookie_banner_translations
    ADD CONSTRAINT cookie_banner_translations_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_sources
    DROP CONSTRAINT access_review_sources_organization_id_fkey;
ALTER TABLE access_review_sources
    ADD CONSTRAINT access_review_sources_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_campaigns
    DROP CONSTRAINT access_review_campaigns_organization_id_fkey;
ALTER TABLE access_review_campaigns
    ADD CONSTRAINT access_review_campaigns_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_entries
    DROP CONSTRAINT access_review_entries_organization_id_fkey;
ALTER TABLE access_review_entries
    ADD CONSTRAINT access_review_entries_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_entry_decision_history
    DROP CONSTRAINT access_review_entry_decision_history_organization_id_fkey;
ALTER TABLE access_review_entry_decision_history
    ADD CONSTRAINT access_review_entry_decision_history_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_campaign_sources
    DROP CONSTRAINT access_review_campaign_sources_organization_id_fkey;
ALTER TABLE access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE access_review_campaign_source_fetch_attempts
    DROP CONSTRAINT access_review_campaign_source_fetch_attemp_organization_id_fkey;
ALTER TABLE access_review_campaign_source_fetch_attempts
    ADD CONSTRAINT access_review_campaign_source_fetch_attemp_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;
