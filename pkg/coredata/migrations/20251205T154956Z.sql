DELETE FROM authz_invitations 
WHERE organization_id NOT IN (SELECT id FROM organizations);

DELETE FROM authz_memberships 
WHERE organization_id NOT IN (SELECT id FROM organizations);

ALTER TABLE authz_invitations ADD CONSTRAINT authz_invitations_organization_id_fkey
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE authz_memberships ADD CONSTRAINT authz_memberships_organization_id_fkey
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;