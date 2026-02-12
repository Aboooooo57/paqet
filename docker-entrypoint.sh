#!/bin/bash
set -e

# Helper function to log
log() {
    echo "[paqet-docker] $1"
}

# Check if we have NET_ADMIN capability
if ! capsh --print | grep -q "cap_net_admin"; then
    log "WARNING: Container lacks NET_ADMIN capability. Iptables rules cannot be applied."
    log "Please run with --cap-add=NET_ADMIN"
fi

# Try to find config file
CONFIG_FILE="config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    # Maybe it's passed as an argument?
    # Iterate args to find -c
    for ((i=1;i<=$#;i++)); do
        if [ "${!i}" = "-c" ]; then
            j=$((i+1))
            CONFIG_FILE="${!j}"
        fi
    done
fi

if [ -f "$CONFIG_FILE" ]; then
    # Try to find role from Env Var first
    if [ -n "$PAQET_ROLE" ]; then
        ROLE="$PAQET_ROLE"
    else
        # Fallback to parsing config file
        ROLE=$(grep 'role:' "$CONFIG_FILE" | awk '{print $2}' | tr -d '"' | tr -d "'")
    fi
    
    if [ "$ROLE" = "server" ]; then
        # Try to find port from Env Var first (Reliable)
        if [ -n "$PAQET_SERVER_ADDR" ]; then
             # Extract port from "IP:PORT" or ":PORT"
             PORT="${PAQET_SERVER_ADDR##*:}"
        else
            # Fallback to parsing config file (Legacy/Hardcoded)
            # Assumes format: addr: ":9999" or addr: "0.0.0.0:9999"
            PORT=$(grep 'addr:' "$CONFIG_FILE" | grep -v '#' | head -n 1 | awk -F':' '{print $NF}' | tr -d '"' | tr -d "'")
        fi
        
        if [ -n "$PORT" ]; then
            log "Detected Server mode on port $PORT. Applying iptables rules..."
            
            # Apply rules (idempotent-ish)
            iptables -t raw -D PREROUTING -p tcp --dport "$PORT" -j NOTRACK 2>/dev/null || true
            iptables -t raw -A PREROUTING -p tcp --dport "$PORT" -j NOTRACK
            
            iptables -t raw -D OUTPUT -p tcp --sport "$PORT" -j NOTRACK 2>/dev/null || true
            iptables -t raw -A OUTPUT -p tcp --sport "$PORT" -j NOTRACK
            
            iptables -t mangle -D OUTPUT -p tcp --sport "$PORT" --tcp-flags RST RST -j DROP 2>/dev/null || true
            iptables -t mangle -A OUTPUT -p tcp --sport "$PORT" --tcp-flags RST RST -j DROP
            
            log "Iptables rules applied successfully."
        else
            log "Could not detect port from $CONFIG_FILE. Skipping iptables setup."
        fi
    fi
else
    log "Config file not found at $CONFIG_FILE. Skipping auto-configuration."
fi

# Execute the command
exec "$@"
