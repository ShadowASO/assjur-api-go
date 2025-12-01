#!/bin/bash
# =============================================================
# check_boot_health.sh
# Verifica estabilidade e falhas silenciosas no Ubuntu Server
# Autor: Aldenor (adaptado por ChatGPT)
# =============================================================

# Cores para saída
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}============================================================="
echo -e "🩺 RELATÓRIO DE INTEGRIDADE DO SISTEMA — $(hostname)"
echo -e "Data: $(date)"
echo -e "=============================================================${NC}\n"

# Últimos boots
echo -e "${YELLOW}➡️  Histórico de boots:${NC}"
last -x | grep -E "reboot|shutdown" | head -n 10
echo ""

# Lista os últimos 5 boots registrados
boots=$(journalctl --list-boots | tail -n 5 | awk '{print $1}')

for b in $boots; do
    echo -e "${CYAN}============================================================="
    echo -e "📅 Boot ID: ${b}"
    echo -e "=============================================================${NC}"

    echo -e "\n${YELLOW}🔸 Eventos críticos:${NC}"
    journalctl -b $b -p 2..3 --no-pager | tail -n 10 || echo "Nenhum evento crítico encontrado."

    echo -e "\n${YELLOW}🔸 Watchdog / Power / Reset:${NC}"
    journalctl -b $b | grep -Ei "watchdog|power|reset|thermal" | tail -n 10 || echo "Nenhum evento de energia encontrado."

    echo -e "\n${YELLOW}🔸 Kernel panic / OOM / travamentos:${NC}"
    journalctl -b $b | grep -Ei "kernel panic|out of memory|oom-killer|BUG:" | tail -n 10 || echo "Nenhum travamento detectado."

    echo -e "\n${YELLOW}🔸 Timeout de rede / serviços pendentes:${NC}"
    journalctl -b $b | grep -Ei "Timeout|networkd-wait-online|failed to start" | tail -n 10 || echo "Sem falhas de rede relevantes."

    echo -e "\n${YELLOW}🔸 Serviços falhando:${NC}"
    systemctl --failed --no-pager || echo "Nenhum serviço com falha."
    echo -e "${CYAN}-------------------------------------------------------------${NC}\n"
done

echo -e "${GREEN}✅ Análise concluída.${NC}"
echo -e "Use 'sudo bash /usr/local/bin/check_boot_health.sh' para execução manual."

