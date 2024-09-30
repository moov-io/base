#!/bin/bash
set -e

chown -R postgres:postgres /opt/moov/

chmod 600 /opt/moov/certs/*.key
chmod 644 /opt/moov/certs/*.crt

chown postgres:postgres /opt/moov/certs/*.key
chown postgres:postgres /opt/moov/certs/*.crt

ls -l /var/lib/postgresql/
