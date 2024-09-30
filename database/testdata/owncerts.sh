#!/bin/bash
set -e

chmod 600 /var/lib/postgresql/*.key
chmod 644 /var/lib/postgresql/*.crt

chown root:root /var/lib/postgresql/*.key
chown root:root /var/lib/postgresql/*.crt

ls -l /var/lib/postgresql/
