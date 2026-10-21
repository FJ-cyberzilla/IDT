#!/data/data/com.termux/files/usr/bin/bash
set -e

# ==========================================
# IDT (Internet Dialer Test) Management Script
# Organization: FJ™ Cybertronic Systems
# ==========================================

# ANSI TrueColor & Styles
COLOR_RESET="\033[0m"
COLOR_BOLD="\033[1m"
COLOR_DIM="\033[2m"
COLOR_GREEN="\033[38;2;52;211;153m"
COLOR_YELLOW="\033[38;2;251;146;60m"
COLOR_RED="\033[38;2;239;68;68m"
COLOR_CYAN="\033[38;2;34;211;238m"
COLOR_BRAND="\033[38;2;255;135;0m"

print_banner() {
    clear
    printf "${COLOR_BRAND}====================================================${COLOR_RESET}\n"
    printf "${COLOR_BRAND}  FJ™ Cybertronic Systems // IDT Management Suite     ${COLOR_RESET}\n"
    printf "${COLOR_BRAND}====================================================${COLOR_RESET}\n\n"
}

print_sep() {
    printf "${COLOR_DIM}----------------------------------------------------${COLOR_RESET}\n"
}

press_enter() {
    echo -e "\n${COLOR_CYAN}Press Enter to return to the management menu...${COLOR_RESET}"
    read -r
}

check_dependencies() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Verifying system dependencies...${COLOR_RESET}\n"
    
    if ! command -v go &> /dev/null; then
        echo -e "${COLOR_YELLOW}[!] Go runtime not found. Attempting installation...${COLOR_RESET}"
        if command -v pkg &> /dev/null; then
            pkg install -y golang git make || { echo -e "${COLOR_RED}[-] Failed to install Go via Termux pkg.${COLOR_RESET}"; return 1; }
        elif command -v apt &> /dev/null; then
            sudo apt update && sudo apt install -y golang git make || { echo -e "${COLOR_RED}[-] Failed to install Go via apt.${COLOR_RESET}"; return 1; }
        else
            echo -e "${COLOR_RED}[-] Error: Go is required. Please install Go manually.${COLOR_RESET}"
            return 1
        fi
    else
        echo -e "${COLOR_GREEN}[✔] Go runtime detected: $(go version)${COLOR_RESET}"
    fi
    
    if ! command -v git &> /dev/null; then
        echo -e "${COLOR_YELLOW}[!] Git not found. Installing git...${COLOR_RESET}"
        pkg install -y git || sudo apt install -y git || true
    else
        echo -e "${COLOR_GREEN}[✔] Git version control detected.${COLOR_RESET}"
    fi
    
    echo -e "\n${COLOR_GREEN}[✔] All system dependencies verified successfully!${COLOR_RESET}"
    press_enter
}

build_app() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Building IDT Binary via Go & Makefile...${COLOR_RESET}\n"
    
    if [ -f "Makefile" ]; then
        make build || { echo -e "${COLOR_RED}[-] Build failed via Makefile.${COLOR_RESET}"; press_enter; return 1; }
    else
        mkdir -p bin
        go build -o bin/idt main.go || { echo -e "${COLOR_RED}[-] Build failed via direct go build.${COLOR_RESET}"; press_enter; return 1; }
    fi
    
    echo -e "\n${COLOR_GREEN}[✔] Build successful! Binary located at: ./bin/idt${COLOR_RESET}"
    press_enter
}

run_tests() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Executing Go Test Suite & Race Detector...${COLOR_RESET}\n"
    
    go test -v -race ./... || { echo -e "\n${COLOR_RED}[-] Test suite encountered failures or race conditions.${COLOR_RESET}"; press_enter; return 1; }
    
    echo -e "\n${COLOR_GREEN}[✔] All tests passed successfully with zero race conditions!${COLOR_RESET}"
    press_enter
}

update_git() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Synchronizing with Remote Git Repository...${COLOR_RESET}\n"
    
    if [ ! -d ".git" ]; then
        echo -e "${COLOR_RED}[-] Error: Not a git repository.${COLOR_RESET}"
        press_enter
        return 1
    fi
    
    git pull origin main || git pull origin master || { echo -e "${COLOR_RED}[-] Git synchronization failed. Check network or local conflicts.${COLOR_RESET}"; press_enter; return 1; }
    
    echo -e "\n${COLOR_GREEN}[✔] Repository successfully updated to the latest remote commit!${COLOR_RESET}"
    press_enter
}

# Main Interactive Menu Loop
while true; do
    print_banner
    echo -e "${COLOR_CYAN}  SELECT MANAGEMENT OPERATION:${COLOR_RESET}\n"
    print_sep
    echo -e "  ${COLOR_BOLD}1)${COLOR_RESET} Check Dependencies & Environment"
    echo -e "  ${COLOR_BOLD}2)${COLOR_RESET} Build IDT Application Binary"
    echo -e "  ${COLOR_BOLD}3)${COLOR_RESET} Run Test Suite & Race Analyzer"
    echo -e "  ${COLOR_BOLD}4)${COLOR_RESET} Update from Remote Git Repository"
    echo -e "  ${COLOR_BOLD}5)${COLOR_RESET} Run IDT Command Center"
    echo -e "  ${COLOR_BOLD}6)${COLOR_RESET} Exit Management Suite"
    print_sep
    echo -ne "\n  ${COLOR_BRAND}Enter choice [1-6]: ${COLOR_RESET}"
    read -r choice

    case "$choice" in
        1) check_dependencies ;;
        2) build_app ;;
        3) run_tests ;;
        4) update_git ;;
        5) 
            if [ -f "./bin/idt" ]; then
                ./bin/idt
            elif [ -f "main.go" ]; then
                go run main.go
            else
                echo -e "${COLOR_RED}[-] Binary not found. Please build first (Option 2).${COLOR_RESET}"
                press_enter
            fi
            ;;
        6)
            print_banner
            echo -e "${COLOR_BRAND} Thank you for using FJ™ Cybertronic Systems IDT Suite.${COLOR_RESET}\n"
            exit 0
            ;;
        *)
            echo -e "\n${COLOR_RED}[!] Invalid choice. Please select between 1 and 6.${COLOR_RESET}"
            sleep 1.5
            ;;
    esac
done
