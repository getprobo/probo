UPDATE
    iam_memberships m
SET
    id = generate_gid(decode_base64_unpadded(o.tenant_id), 39)
FROM
    organizations o
WHERE
    o.id = m.organization_id
    AND extract_entity_type(parse_gid(m.id)) = 38;
