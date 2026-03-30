#!/bin/bash

# Database reset script - drops all tables and runs migrations

# Color definitions
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

cd /Users/dany/Documents/GitHub/oscar || exit 1

# Load environment variables
if [ -f .env ]; then
    source .env
fi

DATABASE_URL="${DATABASE_URL:-postgres://oscar:oscar@localhost:5432/oscar?sslmode=disable}"

echo -e "${YELLOW}🔄 Resetting database...${NC}"
echo -e "${YELLOW}⚠️  This will DROP ALL tables and data!${NC}"
echo -e -n "${YELLOW}Continue? (y/N): ${NC}"
read -r response
if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo -e "${RED}❌ Cancelled${NC}"
    exit 1
fi

# Check if Docker PostgreSQL container is running
if docker ps --format '{{.Names}}' | grep -q "^oscar-postgres$"; then
    echo -e "${BLUE}Using Docker PostgreSQL container${NC}"
    
    docker exec oscar-postgres psql -U oscar -d oscar <<EOF || true
    DROP SCHEMA public CASCADE;
    CREATE SCHEMA public;
    GRANT ALL ON SCHEMA public TO oscar;
    GRANT ALL ON SCHEMA public TO public;
EOF
    echo -e "${GREEN}✓ Tables dropped${NC}"

# Check if psql is available for local PostgreSQL
elif command -v psql &> /dev/null; then
    echo -e "${BLUE}Using local PostgreSQL${NC}"
    
    psql "$DATABASE_URL" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" -c "GRANT ALL ON SCHEMA public TO oscar;" -c "GRANT ALL ON SCHEMA public TO public;" || true
    echo -e "${GREEN}✓ Tables dropped${NC}"
else
    echo -e "${RED}❌ No PostgreSQL connection available${NC}"
    exit 1
fi

# Run migrations
echo -e "${YELLOW}🚀 Running migrations...${NC}"
/Users/dany/go/bin/migrate -path internal/db/migrations -database "$DATABASE_URL" up 2>&1 || true

echo -e "${GREEN}✅ Database reset complete!${NC}"
