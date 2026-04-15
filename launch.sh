#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

SESSION="oscar"

stop() {
    echo -e "${YELLOW}Stopping oscar services...${NC}"

    tmux kill-session -t $SESSION 2>/dev/null && echo -e "${GREEN}oscar session stopped${NC}" || true

    docker stop oscar-minio oscar-mailhog 2>/dev/null && echo -e "${GREEN}Docker containers stopped${NC}" || true
    docker rm oscar-minio oscar-mailhog 2>/dev/null && echo -e "${GREEN}Docker containers removed${NC}" || true

    brew services stop postgresql@16 2>/dev/null && echo -e "${GREEN}PostgreSQL stopped${NC}" || true

    if pgrep -x redis-server > /dev/null 2>&1; then
        pkill -f redis-server && echo -e "${GREEN}Redis stopped${NC}" || true
    fi

    lsof -ti :8080 | xargs kill -9 2>/dev/null && echo -e "${GREEN}Port 8080 cleared${NC}" || true
    lsof -ti :4321 | xargs kill -9 2>/dev/null && echo -e "${GREEN}Port 4321 cleared${NC}" || true

    echo ""
    echo -e "${GREEN}All services stopped.${NC}"
}

reset_db() {
    echo -e "${YELLOW}Resetting database...${NC}"

    source .env 2>/dev/null || true

    if [ -z "$DATABASE_URL" ]; then
        echo -e "${RED}DATABASE_URL not found in .env${NC}"
        exit 1
    fi

    DB_NAME=$(echo "$DATABASE_URL" | sed 's|.*/||' | sed 's|\?.*||')

    echo -e "${YELLOW}Target database: ${DB_NAME}${NC}"

    read -p "This will DELETE ALL DATA in the database. Are you sure? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Aborted${NC}"
        exit 0
    fi

    echo -e "${YELLOW}Truncating all tables...${NC}"

    psql "$DATABASE_URL" <<-EOSQL
		DO \$\$
		DECLARE
		    r RECORD;
		BEGIN
		    FOR r IN (
		        SELECT tablename FROM pg_tables
		        WHERE schemaname = 'public'
		        AND tablename NOT IN ('schema_migrations', 'spatial_ref_sys')
		    ) LOOP
		        EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
		    END LOOP;
		END
		\$\$;
EOSQL

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Database reset complete${NC}"
    else
        echo -e "${RED}Database reset failed${NC}"
        exit 1
    fi
}

start() {
    echo -e "${YELLOW}Cleaning up existing processes...${NC}"
    pkill -f "oscar" 2>/dev/null || true
    pkill -f "postgres" 2>/dev/null || true
    pkill -f "redis" 2>/dev/null || true
    pkill -f "ngrok" 2>/dev/null || true
    docker stop oscar-minio oscar-mailhog 2>/dev/null || true
    docker rm oscar-minio oscar-mailhog 2>/dev/null || true
    sleep 2
    echo -e "${GREEN}Cleanup complete${NC}"

    cd /Users/dany/Documents/GitHub/oscar

    current_branch=$(git branch --show-current)
    echo -e "${BLUE}Current branch: ${current_branch}${NC}"

    echo "Fetching latest changes..."
    git fetch origin
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Fetch successful${NC}"
        updates=$(git rev-list HEAD..origin/$current_branch --count 2>/dev/null)
        if [ "$updates" -gt 0 ] 2>/dev/null; then
            echo -e "${YELLOW}Found ${updates} new commit(s). Pulling changes...${NC}"
            git pull origin $current_branch
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}Pull successful${NC}"
                echo -e "${BLUE}Recent commits:${NC}"
                git log --oneline -5
            else
                echo -e "${RED}Pull failed. Please check for conflicts.${NC}"
            fi
        else
            echo -e "${GREEN}Already up to date${NC}"
        fi
    else
        echo -e "${RED}Fetch failed${NC}"
    fi

    echo -e "${YELLOW}Running database migrations...${NC}"
    source .env && /Users/dany/go/bin/migrate -path internal/db/migrations -database "$DATABASE_URL" up 2>&1 | grep -i error || echo -e "${GREEN}Migrations complete${NC}"

    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    cd web && npm i 2>&1 | grep -iE "(error|ERR)" || echo -e "${GREEN}Dependencies installed${NC}"
    cd ..

    tmux kill-session -t $SESSION 2>/dev/null || true

    tmux new-session -d -s $SESSION -n 'main'

    tmux split-window -h -t $SESSION

    tmux split-window -h -t $SESSION

    tmux select-layout -t $SESSION even-vertical

    tmux send-keys -t $SESSION:0.0 "clear && echo -e '${PURPLE}=== OSCAR API ===${NC}' && echo 'Waiting for MinIO...' && while ! curl -s --max-time 1 http://localhost:9000/minio/health/live > /dev/null 2>&1; do sleep 1; done && while ! /usr/local/bin/redis-cli ping 2>/dev/null | grep -q PONG; do sleep 1; done && echo 'Ready! Starting API...' && cd /Users/dany/Documents/GitHub/oscar && lsof -ti :8080 | xargs kill -9 2>/dev/null; source .env && make dev" C-m

    tmux send-keys -t $SESSION:0.1 "clear && echo -e '${CYAN}=== ASTRO FRONTEND ===${NC}' && cd /Users/dany/Documents/GitHub/oscar/web && npx astro dev" C-m

    tmux send-keys -t $SESSION:0.2 "clear && echo -e '${GREEN}=== NGROK ===${NC}' && (ngrok http 8080 --log=stdout --log-level=info 2>/dev/null || echo 'ngrok not available')" C-m

    tmux new-window -t $SESSION -n 'services'

    tmux send-keys -t $SESSION:1 "clear && echo '' && echo 'Starting services...' && brew services start postgresql@16 && docker run -d --name oscar-minio -p 9000:9000 -p 9001:9001 -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin minio/minio server /data --console-address ':9001' && docker run -d --name oscar-mailhog -p 1025:1025 -p 8025:8020 mailhog/mailhog && /usr/local/bin/redis-server --daemonize yes && echo 'All services started' && echo ''" C-m

    tmux select-window -t $SESSION:0

    echo ""
    echo -e "${GREEN}All services started in tmux sessions${NC}"
    echo -e "${YELLOW}Git:${NC}"
    echo "  - oscar: $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
    echo ""
    echo -e "${YELLOW}Sessions:${NC}"
    echo "  tmux attach -t oscar        -> Backend + Frontend + Ngrok"
    echo "  tmux attach -t oscar-infra  -> Postgres, MinIO, Mailhog, Redis"
    echo ""
    echo -e "${YELLOW}URLs:${NC}"
    echo "  API:       http://localhost:8080"
    echo "  Frontend:  http://localhost:4321"
    echo "  MinIO:     http://localhost:9001"
    echo "  Mailhog:   http://localhost:8025"
    echo ""
    echo -e "${YELLOW}Navigation:${NC}"
    echo "  Ctrl+B, 0 = oscar + Astro (main)"
    echo "  Ctrl+B, 1 = Services (Postgres, MinIO, Mailhog)"
    echo "  Ctrl+B, D = Detach"
    echo ""
    echo -e "${GREEN}Attaching to tmux session...${NC}"
    sleep 1
    tmux attach -t $SESSION
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    reset-db)
        reset_db
        ;;
    *)
        echo ""
        echo -e "${YELLOW}oscar Script${NC}"
        echo ""
        echo "Usage: ./launch.sh [command]"
        echo ""
        echo "Commands:"
        echo "  start     Start all services (API, Frontend, Ngrok, Postgres, MinIO, Mailhog, Redis)"
        echo "  stop      Stop all services"
        echo "  reset-db  Truncate all tables in the database (requires confirmation)"
        echo ""
        echo "No argument shows this help."
        echo ""
        ;;
esac
