#!/bin/sh
# Backup harian Postgres: pg_dump format custom (-Fc), retensi 14 hari (4.5).
# Dijadwalkan di host (cron/systemd timer), bukan di dalam container:
#   BACKUP_DIR=/var/backups/stockopname ./ops/backup/backup.sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
COMPOSE="${COMPOSE:-docker compose}"

mkdir -p "$BACKUP_DIR"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="$BACKUP_DIR/stockopname_${STAMP}.dump"

# pg_dump jalan di dalam container postgres; kredensial dari env container.
$COMPOSE exec -T postgres sh -c \
  'pg_dump -Fc -U "$POSTGRES_USER" -d "$POSTGRES_DB"' > "$OUT.tmp"

# Validasi: file tidak kosong dan katalognya terbaca oleh pg_restore --list.
test -s "$OUT.tmp"
pg_restore --list "$OUT.tmp" >/dev/null
mv "$OUT.tmp" "$OUT"
echo "backup ok: $OUT ($(wc -c < "$OUT") bytes)"

# Retensi: hapus dump yang lebih tua dari RETENTION_DAYS.
find "$BACKUP_DIR" -name 'stockopname_*.dump' -mtime +"$RETENTION_DAYS" -print -delete
