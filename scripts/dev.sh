#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${YELLOW}Cleaning up existing processes...${NC}"
pkill -f "oscar" 2>/dev/null
pkill -f "postgres" 2>/dev/null
pkill -f "redis" 2>/dev/null
pkill -f "tmux" 2>/dev/null
pkill -f "ngrok" 2>/dev/null
docker stop oscar-mailhog 2>/dev/null
docker rm oscar-mailhog 2>/dev/null
sleep 2
echo -e "${GREEN}✓ Killed existing processes${NC}"

cd /Users/dany/Documents/GitHub/oscar || exit 1

current_branch=$(git branch --show-current)
echo -e "${BLUE}Current branch: ${current_branch}${NC}"

echo -e "Fetching latest changes..."
git fetch origin
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Fetch successful${NC}"

    updates=$(git rev-list HEAD..origin/$current_branch --count 2>/dev/null)
    if [ "$updates" -gt 0 ] 2>/dev/null; then
        echo -e "${YELLOW}Found ${updates} new commit(s). Pulling changes...${NC}"

        git pull origin $current_branch
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Pull successful${NC}"

            echo -e "${BLUE}Recent commits:${NC}"
            git log --oneline -5
        else
            echo -e "${RED}✗ Pull failed. Please check for conflicts.${NC}"
        fi
    else
        echo -e "${GREEN}✓ Already up to date${NC}"
    fi
else
    echo -e "${RED}✗ Fetch failed${NC}"
fi

echo -e "${YELLOW}Running database migrations...${NC}"
source .env && /Users/dany/go/bin/migrate -path internal/db/migrations -database "$DATABASE_URL" up 2>&1 | grep -i error || echo -e "${GREEN}✓ Migrations complete${NC}"

echo -e "${YELLOW}Installing frontend dependencies...${NC}"
cd web && npm i 2>&1 | grep -iE "(error|ERR)" || echo -e "${GREEN}✓ Dependencies installed${NC}"
cd ..

SESSION_NAME="oscar"

tmux kill-session -t $SESSION_NAME 2>/dev/null

tmux new-session -d -s $SESSION_NAME -n 'main'

tmux split-window -h -t $SESSION_NAME

tmux send-keys -t $SESSION_NAME:0.0 "cd /Users/dany/Documents/GitHub/oscar && lsof -ti :8080 | xargs kill -9 2>/dev/null; clear && echo -e '${PURPLE}=== OSCAR APP ===${NC}' && source .env && make dev" C-m

tmux send-keys -t $SESSION_NAME:0.1 "cd /Users/dany/Documents/GitHub/oscar/web && clear && echo -e '${CYAN}=== ASTRO FRONTEND ===${NC}' && npx astro dev" C-m

tmux new-window -t $SESSION_NAME:1 -n 'services'

tmux split-window -h -t $SESSION_NAME:1

tmux send-keys -t $SESSION_NAME:1.0 "clear && echo -e '${PURPLE}=== POSTGRESQL ===${NC}' && brew services start postgresql@16" C-m

tmux send-keys -t $SESSION_NAME:1.1 "clear && echo -e '${PURPLE}=== MAILHOG ===${NC}' && (docker run -d --name oscar-mailhog -p 1025:1025 -p 8025:8020 mailhog/mailhog 2>/dev/null || echo 'Docker not available')" C-m

tmux new-window -t $SESSION_NAME:2 -n 'tools'

tmux split-window -h -t $SESSION_NAME:2

tmux send-keys -t $SESSION_NAME:2.0 "clear && echo -e '${YELLOW}=== REDIS ===${NC}' && /usr/local/bin/redis-server --daemonize yes" C-m

tmux send-keys -t $SESSION_NAME:2.1 "clear && echo -e '${RED}=== NGROK ===${NC}' && (ngrok http 8080 --log=stdout --log-level=info 2>/dev/null || echo 'ngrok not available')" C-m

tmux select-window -t $SESSION_NAME:0

echo ""
echo -e "${GREEN}✓ All services started in tmux session${NC}"
echo -e "${YELLOW}Git:${NC}"
echo "  - oscar: $(cd /Users/dany/Documents/GitHub/oscar && git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
echo ""
echo -e "${YELLOW}Services:${NC}"
echo "  - oscar API: http://localhost:8080"
echo "  - Astro Frontend: http://localhost:4321"
echo "  - Mailhog: http://localhost:8025"
echo "  - Ngrok: http://localhost:4040"
echo ""
echo -e "${YELLOW}Navigation:${NC}"
echo "  - Ctrl+B, 0 = oscar + Astro (main)"
echo "  - Ctrl+B, 1 = Services (Postgres, Mailhog)"
echo "  - Ctrl+B, 2 = Tools (Redis, Ngrok)"
echo "  - Ctrl+B, D = Detach"
echo "  - Ctrl+B, [ = Scroll mode (q to exit)"
echo ""
echo -e "${GREEN}Attaching to tmux session in 2 seconds...${NC}"
sleep 2
tmux attach -t $SESSION_NAME
