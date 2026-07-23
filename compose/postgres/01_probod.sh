#!/bin/sh

set -eu

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" <<-EOF
CREATE USER probod;
ALTER USER probod WITH SUPERUSER;
ALTER USER probod PASSWORD 'probod';
CREATE DATABASE probod;
GRANT ALL PRIVILEGES ON DATABASE probod TO probod;
CREATE DATABASE probod_test;
GRANT ALL PRIVILEGES ON DATABASE probod_test TO probod;
EOF

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d probod <<-EOF
ALTER SCHEMA public OWNER TO probod;
GRANT ALL ON SCHEMA public TO probod;
ALTER DATABASE probod SET probo.trust_center_base_domain TO 'probopage.localhost';
EOF

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d probod_test <<-EOF
ALTER SCHEMA public OWNER TO probod;
GRANT ALL ON SCHEMA public TO probod;
ALTER DATABASE probod_test SET probo.trust_center_base_domain TO 'probopage.localhost';
EOF
