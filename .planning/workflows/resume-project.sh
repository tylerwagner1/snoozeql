#!/bin/bash
set -e

# Resume Project Workflow
# Restores complete project context and presents status

cd "$(dirname "$0")/../.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}╔$(printf '═%.0s' {1..62})╗${NC}"
    echo -e "${BLUE}║${NC} $(printf '%-60s' "$1") ${BLUE}║${NC}"
    echo -e "${BLUE}╚$(printf '═%.0s' {1..62})╝${NC}"
}

print_section() {
    echo -e "${BLUE}├$(printf '─%.0s' {1..62})┤${NC}"
    echo -e "${BLUE}│${NC} $(printf '%-60s' "$1") ${BLUE}│${NC}"
    echo -e "${BLUE}├$(printf '─%.0s' {1..62})┤${NC}"
}

# Step 1: Detect existing project
detect_existing_project() {
    echo "🔍 Detecting existing project..."
    
    if [ -f ".planning/STATE.md" ]; then
        echo "✓ STATE.md exists"
        return 0
    elif [ -f ".planning/ROADMAP.md" ] || [ -f ".planning/PROJECT.md" ]; then
        echo "⚠️  STATE.md missing but artifacts exist"
        return 1
    else
        echo "❌ No planning directory or artifacts found"
        return 2
    fi
}

# Step 2: Load state from files
load_state() {
    echo -e "\n${BLUE}📚 Loading project state...${NC}\n"
    
    if [ ! -f ".planning/STATE.md" ]; then
        echo -e "${YELLOW}STATE.md missing. Reconstructing from artifacts...${NC}\n"
        reconstruct_state
        return $?
    fi
    
    echo "=== STATE.md ==="
    cat .planning/STATE.md
    echo ""
    
    if [ -f ".planning/PROJECT.md" ]; then
        echo "=== PROJECT.md ==="
        cat .planning/PROJECT.md
        echo ""
    fi
}

# Reconstruct STATE.md from available artifacts
reconstruct_state() {
    if [ ! -f ".planning/PROJECT.md" ]; then
        echo "❌ Cannot reconstruct STATE.md without PROJECT.md"
        return 1
    fi
    
    echo "Reconstructing STATE.md..."
    
    local project_what=""
    local project_value=""
    local current_phase=""
    local phases_count=0
    local plans_in_phase=0
    local completed_plans=0
    local last_activity=""
    local pending_todos=0
    local blockers=""
    local session_continuity=""
    
    # Extract from PROJECT.md
    if [ -f ".planning/PROJECT.md" ]; then
        project_what=$(grep -A2 "What This Is" .planning/PROJECT.md | tail -1 | sed 's/^[[:space:]]*//')
        project_value=$(grep -i "core\|value" .planning/PROJECT.md | head -1 | sed 's/^[[:space:]]*//')
    fi
    
    # Extract from ROADMAP.md
    if [ -f ".planning/ROADMAP.md" ]; then
        phases_count=$(grep -c "^### Phase" .planning/ROADMAP.md 2>/dev/null || echo "0")
        current_phase=$(grep "^\*\*Phase.*\*\*" .planning/ROADMAP.md 2>/dev/null | head -1 | sed 's/.*Phase \([0-9]*\).*/\1/')
        [ -z "$current_phase" ] && current_phase="1"
    fi
    
    # Count completed plans
    plans_in_phase=$(ls .planning/phases/*/ 2>/dev/null | wc -l)
    completed_plans=$(ls .planning/phases/*/*-SUMMARY.md 2>/dev/null | wc -l)
    
    # Count pending todos
    if [ -d ".planning/todos/pending/" ]; then
        pending_todos=$(ls .planning/todos/pending/*.md 2>/dev/null | wc -l)
    fi
    
    # Check for incomplete work
    local continue_here=$(ls .planning/phases/*/.continue-here*.md 2>/dev/null | head -1)
    local interrupted_agent=""
    
    if [ -f ".planning/current-agent-id.txt" ] && [ -s ".planning/current-agent-id.txt" ]; then
        interrupted_agent=$(cat .planning/current-agent-id.txt | tr -d '\n')
    fi
    
    # Generate STATE.md content
    cat > .planning/STATE.md << EOF
# Project State

## Project Reference
**Project:** ${project_what:-Unknown Project}
**Core Value:** ${project_value:-Building a project}

## Current Position
**Phase:** ${current_phase:-1} of ${phases_count:-1}
**Plans Completed:** ${completed_plans:-0} of ${plans_in_phase:-0}

## Progress
$(generate_progress_bar "$completed_plans" "$plans_in_phase")

## Recent Decisions
No decisions recorded yet.

## Pending Todos
${pending_todos} pending ideas captured during sessions.

## Blockers/Concerns
No blockers carried forward.

## Session Continuity
Last session: $(date '+%Y-%m-%d %H:%M:%S')
Stopped at: Project state reconstructed from artifacts
Resume file: None
EOF
    
    echo "✓ STATE.md reconstructed successfully"
}

generate_progress_bar() {
    local completed=$1
    local total=$2
    [ "$total" -eq 0 ] && total=1
    
    local percentage=$((completed * 100 / total))
    local filled=$((percentage / 10))
    local empty=$((10 - filled))
    
    printf "["
    for ((i=0; i<filled; i++)); do printf "█"; done
    for ((i=0; i<empty; i++)); do printf "░"; done
    printf "] %d%%" "$percentage"
}

# Step 3: Check for incomplete work
check_incomplete_work() {
    echo -e "\n${BLUE}🔍 Checking for incomplete work...${NC}\n"
    
    local found_incomplete=false
    
    # Check for continue-here files
    local continue_files=$(find .planning/phases -name ".continue-here*.md" 2>/dev/null)
    if [ -n "$continue_files" ]; then
        echo -e "${YELLOW}⚠️  Found mid-plan checkpoint(s):${NC}"
        echo "$continue_files" | while read -r file; do
            echo "  - $file"
        done
        found_incomplete=true
    fi
    
    # Check for plans without summaries
    echo ""
    echo "Checking for incomplete plan executions..."
    local incompleted=0
    for plan in $(find .planning/phases -name "*-PLAN.md" 2>/dev/null); do
        local summary="${plan/PLAN/SUMMARY}"
        if [ ! -f "$summary" ]; then
            echo "  - $plan (no SUMMARY)"
            incompleted=$((incompleted + 1))
        fi
    done
    
    if [ "$incompleted" -gt 0 ]; then
        echo -e "${YELLOW}Found $incompleted incomplete plan(s)${NC}"
        found_incomplete=true
    fi
    
    # Check for interrupted agent
    if [ -f ".planning/current-agent-id.txt" ] && [ -s ".planning/current-agent-id.txt" ]; then
        local agent_id=$(cat .planning/current-agent-id.txt | tr -d '\n')
        echo -e "\n${YELLOW}⚠️  Interrupted agent detected:${NC}"
        echo "  - Agent ID: $agent_id"
        
        if [ -f ".planning/agent-history.json" ]; then
            echo "  - Task: $(cat .planning/agent-history.json | grep -o '"task":"[^"]*"' | head -1 | cut -d'"' -f4)"
        fi
        
        found_incomplete=true
    fi
    
    return 0
}

# Step 4: Present status
present_status() {
    echo ""
    print_header "  PROJECT STATUS  "
    
    local project_what=""
    if [ -f ".planning/PROJECT.md" ]; then
        project_what=$(grep -A2 "What This Is" .planning/PROJECT.md | tail -1 | sed 's/^[[:space:]]*//' | sed 's/\*\*//g')
    fi
    [ -z "$project_what" ] && project_what="Unknown Project"
    
    local current_phase="1"
    local phases_count="1"
    if [ -f ".planning/ROADMAP.md" ]; then
        phases_count=$(grep -c "^### Phase" .planning/ROADMAP.md 2>/dev/null || echo "1")
        current_phase=$(grep "^\*\*Phase.*\*\*" .planning/ROADMAP.md 2>/dev/null | head -1 | sed 's/.*Phase \([0-9]*\).*/\1/')
        [ -z "$current_phase" ] && current_phase="1"
    fi
    
    local completed_plans=$(ls .planning/phases/*/*-SUMMARY.md 2>/dev/null | wc -l)
    local total_plans=$(find .planning/phases -name "*-PLAN.md" 2>/dev/null | wc -l)
    
    echo "Building: $project_what"
    echo ""
    echo "Phase: $current_phase of $phases_count"
    echo "Plans: $completed_plans of $total_plans complete"
    echo "Progress: $(generate_progress_bar $completed_plans $total_plans)"
    echo ""
    
    # Show last activity
    if [ -f ".planning/STATE.md" ]; then
        local last_activity=$(grep "Last session" .planning/STATE.md | head -1 | sed 's/Last session: //')
        echo "Last activity: $last_activity"
    fi
    
    echo ""
    echo "──────────────────────────────────────────────────────────────────────"
    
    # Show incomplete work if found
    local continue_files=$(find .planning/phases -name ".continue-here*.md" 2>/dev/null)
    if [ -n "$continue_files" ]; then
        echo ""
        echo -e "${YELLOW}⚠️  Incomplete work detected:${NC}"
        echo "$continue_files" | while read -r file; do
            echo "  • $file"
        done
        echo ""
    fi
    
    # Check for interrupted agent
    if [ -f ".planning/current-agent-id.txt" ] && [ -s ".planning/current-agent-id.txt" ]; then
        local agent_id=$(cat .planning/current-agent-id.txt | tr -d '\n')
        echo ""
        echo -e "${YELLOW}⚠️  Interrupted agent detected:${NC}"
        echo "  Agent ID: $agent_id"
        echo ""
        echo "  Resume with: Task tool (resume parameter with agent ID)"
        echo ""
    fi
    
    # Show pending todos
    local pending_todos=0
    if [ -d ".planning/todos/pending/" ]; then
        pending_todos=$(ls .planning/todos/pending/*.md 2>/dev/null | wc -l)
    fi
    if [ "$pending_todos" -gt 0 ]; then
        echo "📋 $pending_todos pending todos — /gsd-check-todos to review"
        echo ""
    fi
    
    # Check blockers
    if [ -f ".planning/STATE.md" ]; then
        local blockers=$(grep -A10 "## Blockers" .planning/STATE.md | tail -n +2 | grep -v "^## " | grep -v "^$" | head -5)
        if [ -n "$blockers" ]; then
            echo -e "${YELLOW}⚠️  Carried concerns:${NC}"
            echo "$blockers" | while read -r line; do
                [ -n "$line" ] && echo "  - $line"
            done
            echo ""
        fi
    fi
    
    echo "──────────────────────────────────────────────────────────────────────"
}

# Step 5: Determine next action
determine_next_action() {
    echo ""
    print_header "  NEXT ACTION DETECTION  "
    echo ""
    
    # Check for interrupted agent
    if [ -f ".planning/current-agent-id.txt" ] && [ -s ".planning/current-agent-id.txt" ]; then
        echo "Primary: Resume interrupted agent"
        echo "Option: Start fresh"
        echo ""
        echo "Command: Task tool with resume parameter"
        echo "  → /gsd-resume-agent"
        return 0
    fi
    
    # Check for continue-here files
    local continue_file=$(find .planning/phases -name ".continue-here*.md" 2>/dev/null | head -1)
    if [ -n "$continue_file" ]; then
        echo "Primary: Resume from checkpoint"
        echo "Option: Start fresh"
        echo ""
        echo "Command: Resume from $continue_file"
        echo "  → /gsd-resume-checkpoint"
        return 0
    fi
    
    # Check for incomplete plans
    local incomplete_plan=$(find .planning/phases -name "*-PLAN.md" 2>/dev/null | while read -r plan; do
        local summary="${plan/PLAN/SUMMARY}"
        [ ! -f "$summary" ] && echo "$plan"
    done | head -1)
    
    if [ -n "$incomplete_plan" ]; then
        echo "Primary: Complete incomplete plan"
        echo "Option: Abandon and move on"
        echo ""
        echo "Command: Complete plan - $incomplete_plan"
        echo "  → /gsd-complete-incomplete"
        return 0
    fi
    
    # Check if phase has unexecuted plans
    local current_phase_dir=$(ls -d .planning/phases/* 2>/dev/null | tail -1)
    if [ -n "$current_phase_dir" ]; then
        local phase_name=$(basename "$current_phase_dir")
        local phase_plans=$(find "$current_phase_dir" -name "*-PLAN.md" 2>/dev/null | wc -l)
        local phase_completed=$(find "$current_phase_dir" -name "*-SUMMARY.md" 2>/dev/null | wc -l)
        
        if [ "$phase_completed" -lt "$phase_plans" ]; then
            echo "Primary: Execute next plan"
            echo "Option: Review plan first"
            echo ""
            echo "Command: Execute phase '$phase_name'"
            echo "  → /gsd-execute-phase $phase_name"
            return 0
        fi
    fi
    
    # Check if all phases completed
    local phases_count=$(ls -d .planning/phases/* 2>/dev/null | wc -l)
    local completed_phases=$(ls -d .planning/phases/* 2>/dev/null | while read -r phase; do
        local plans=$(find "$phase" -name "*-PLAN.md" 2>/dev/null | wc -l)
        local summaries=$(find "$phase" -name "*-SUMMARY.md" 2>/dev/null | wc -l)
        [ "$plans" -eq "$summaries" ] && echo "$phase"
    done | wc -l)
    
    if [ "$completed_phases" -ge "$phases_count" ] && [ "$phases_count" -gt 0 ]; then
        echo "Primary: Transition to next phase"
        echo "Option: Review completed work"
        echo ""
        echo "Command: Phase complete - moving to next"
        echo "  → /gsd-transition-phase"
        return 0
    fi
    
    # Check if phase needs planning
    echo "Primary: Plan phase"
    echo "Option: Discuss context first"
    echo ""
    echo "Command: Plan current phase"
    echo "  → /gsd-plan-phase"
    return 0
}

# Step 6: Offer options
offer_options() {
    echo ""
    print_header "  CHOOSE NEXT ACTION  "
    echo ""
    echo "What would you like to do?"
    echo ""
    echo "1. [Primary action - see above]"
    echo "2. Review current status"
    echo "3. Check pending todos"
    echo "4. Review project brief"
    echo "5. Something else"
    echo ""
}

# Step 7: Update session continuity
update_session() {
    echo ""
    echo "Updating session continuity..."
    
    if [ -f ".planning/STATE.md" ]; then
        sed -i.bak "/^## Session Continuity/,/Resume file:/{
            s/^Last session:.*$/Last session: $(date '+%Y-%m-%d %H:%M:%S')/
           s/^Stopped at:.*$/Stopped at: Session resumed at $(date '+%Y-%m-%d %H:%M:%S')/
        }" .planning/STATE.md 2>/dev/null || true
        rm -f .planning/STATE.md.bak 2>/dev/null || true
    fi
    
    echo "✓ Session updated"
}

# Main workflow
main() {
    echo ""
    print_header "  RESUME PROJECT WORKFLOW  "
    echo ""
    
    # Step 1: Detect existing project
    detect_existing_project
    local project_status=$?
    
    case $project_status in
        0)
            # STATE.md exists, proceed normally
            ;;
        1)
            # STATE.md missing but artifacts exist
            echo ""
            echo -e "${YELLOW}STATE.md missing. Reconstructing from artifacts...${NC}"
            reconstruct_state
            ;;
        2)
            # No planning directory
            echo -e "${RED}❌ This appears to be a new project${NC}"
            echo ""
            echo "No .planning/ directory or artifacts found."
            echo "Run /gsd-new-project to start a new project."
            exit 0
            ;;
    esac
    
    # Step 2: Load state
    load_state
    
    # Step 3: Check incomplete work
    check_incomplete_work
    
    # Step 4: Present status
    present_status
    
    # Step 5: Determine next action
    determine_next_action
    
    # Step 6: Offer options
    offer_options
    
    # Step 7: Update session
    update_session
    
    echo ""
    print_header "  END OF RESUME WORKFLOW  "
    echo ""
    echo "To proceed, enter your choice (1-5) or describe what you'd like to do."
    echo ""
}

main "$@"
