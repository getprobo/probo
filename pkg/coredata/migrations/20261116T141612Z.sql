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

-- Free the relation name assets for the resource inventory, and give the
-- register's tables, columns, and constraints their final names.
-- PostgreSQL keeps constraint names when a table is renamed.
-- On PostgreSQL 18 a primary-key index shares its constraint's name,
-- so renaming a primary key renames that index as well.
ALTER TABLE assets RENAME TO assets_registry;

ALTER TABLE assets_registry RENAME CONSTRAINT assets_amount_not_null TO assets_registry_amount_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_asset_type_not_null TO assets_registry_registry_asset_type_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_created_at_not_null TO assets_registry_created_at_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_id_not_null TO assets_registry_id_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_name_not_null TO assets_registry_name_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_organization_id_fkey TO assets_registry_organization_id_fkey;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_organization_id_not_null TO assets_registry_organization_id_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_owner_profile_id_fkey TO assets_registry_owner_profile_id_fkey;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_owner_profile_id_not_null TO assets_registry_owner_profile_id_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_pkey TO assets_registry_pkey;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_tenant_id_not_null TO assets_registry_tenant_id_not_null;
ALTER TABLE assets_registry RENAME CONSTRAINT assets_updated_at_not_null TO assets_registry_updated_at_not_null;

ALTER TABLE generated_documents
    RENAME CONSTRAINT generated_documents_asset_list_document_id_fkey
    TO generated_documents_registry_asset_list_document_id_fkey;

ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_asset_id_fkey TO registry_asset_third_parties_registry_asset_id_fkey;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_asset_id_not_null TO registry_asset_third_parties_registry_asset_id_not_null;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_created_at_not_null TO registry_asset_third_parties_created_at_not_null;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_organization_id_not_null TO registry_asset_third_parties_organization_id_not_null;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_pkey TO registry_asset_third_parties_pkey;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_tenant_id_not_null TO registry_asset_third_parties_tenant_id_not_null;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_vendor_id_fkey TO registry_asset_third_parties_third_party_id_fkey;
ALTER TABLE asset_third_parties
    RENAME CONSTRAINT asset_vendors_vendor_id_not_null TO registry_asset_third_parties_third_party_id_not_null;

ALTER TABLE asset_third_parties RENAME TO registry_asset_third_parties;

ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_asset_id_fkey TO business_function_registry_assets_registry_asset_id_fkey;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_asset_id_not_null TO business_function_registry_assets_registry_asset_id_not_null;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_business_function_id_fkey TO business_function_registry_assets_business_function_id_fkey;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_business_function_id_not_null TO business_function_registry_assets_business_function_id_not_null;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_created_at_not_null TO business_function_registry_assets_created_at_not_null;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_organization_id_fkey TO business_function_registry_assets_organization_id_fkey;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_organization_id_not_null TO business_function_registry_assets_organization_id_not_null;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_pkey TO business_function_registry_assets_pkey;
ALTER TABLE business_function_assets
    RENAME CONSTRAINT business_function_assets_tenant_id_not_null TO business_function_registry_assets_tenant_id_not_null;

ALTER TABLE business_function_assets RENAME TO business_function_registry_assets;

ALTER TABLE assets_registry
    RENAME COLUMN asset_type TO registry_asset_type;

ALTER TABLE registry_asset_third_parties
    RENAME COLUMN asset_id TO registry_asset_id;

ALTER TABLE business_function_registry_assets
    RENAME COLUMN asset_id TO registry_asset_id;
