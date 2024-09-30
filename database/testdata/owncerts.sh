#!/bin/bash
set -e

chmod 600 /var/lib/postgresql/*.key
chmod 644 /var/lib/postgresql/*.crt

chown postgres:postgres /var/lib/postgresql/*.key
chown postgres:postgres /var/lib/postgresql/*.crt

ls -l /var/lib/postgresql/
