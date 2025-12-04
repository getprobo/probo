UPDATE
    reports
SET
    organization_id = a.organization_id
FROM
    audits a
WHERE
    reports.id = a.report_id;
