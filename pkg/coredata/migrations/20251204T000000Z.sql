UPDATE
    reports
SET
    organization_id = a.organization_id
FROM
    audits a
WHERE
    reports.tenant_id = a.tenant_id;
