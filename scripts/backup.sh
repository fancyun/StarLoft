#!/usr/bin/env bash
# 本地定期备份脚本：MySQL 逻辑备份（mysqldump）+ Redis RDB 快照
# 用途：配合宿主 cron 定时执行，备份产物存于项目根备份目录（默认 ./backups，可被 .dockerignore 排除）。
# 用法：
#   ./scripts/backup.sh                    # 立即执行一次备份
#   ./scripts/backup.sh --retention 7      # 仅保留 7 天，更早的备份目录自动删除
# 宿主 cron 示例（每天凌晨 3:00 执行，保留最近 7 天）：
#   0 3 * * * /path/to/StarLoft/scripts/backup.sh --retention 7 >> /path/to/StarLoft/logs/backup.log 2>&1
set -euo pipefail

# 项目根目录（本脚本所在目录的上一级）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# 读取 .env（仅取出所需变量，不导入全部避免污染环境变量）
ENV_FILE="$PROJECT_ROOT/.env"
if [ ! -f "$ENV_FILE" ]; then
  echo "错误: 未找到 $ENV_FILE" >&2
  exit 1
fi

DB_PASSWORD=""
REDIS_PASSWORD=""
while IFS='=' read -r key value; do
  key="$(echo "$key" | tr -d '[:space:]')"
  case "$key" in
    DB_PASSWORD) DB_PASSWORD="${value%\r}" ;;
    REDIS_PASSWORD) REDIS_PASSWORD="${value%\r}" ;;
  esac
done < "$ENV_FILE"

# 备份目录与保留天数
BACKUP_ROOT="${BACKUP_ROOT:-$PROJECT_ROOT/backups}"
RETENTION_DAYS=7
while [ "$#" -gt 0 ]; do
  case "$1" in
    --retention) RETENTION_DAYS="$2"; shift 2 ;;
    *) shift ;;
  esac
done

DEST="$BACKUP_ROOT/$(date +%Y%m%d)"
mkdir -p "$DEST"

MYSQL_DBS=("starloft_sys" "starloft_kyc")

echo "==> 开始备份 -> $DEST"
for db in "${MYSQL_DBS[@]}"; do
  echo "  dump MySQL 库: $db"
  docker compose exec -T mysql \
    mysqldump -uroot -p"$DB_PASSWORD" --single-transaction --routines --triggers \
    --databases "$db" > "$DEST/${db}_$(date +%H%M%S).sql"
done

echo "  Redis RDB 快照"
docker compose exec -T redis redis-cli --no-auth-warning -a "$REDIS_PASSWORD" SAVE
docker compose cp redis:/data/dump.rdb "$DEST/redis_dump_$(date +%H%M%S).rdb"

echo "  清理超过 ${RETENTION_DAYS} 天的备份"
find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -mtime "+${RETENTION_DAYS}" -exec rm -rf {} +

echo "==> 备份完成: $DEST"