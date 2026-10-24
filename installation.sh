#!/usr/bin/env bash
set -e

# ==========================================
# IDT (Internet Dialer Test) Management Suite
# Organization: FJ™ Cybertronic Systems
# Target: Termux, WSL2, Linux, macOS
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

# ==========================================
# Environment Detection
# ==========================================
is_termux() {
    [ -n "${TERMUX_VERSION:-}" ] || [[ "${PREFIX:-}" == *com.termux* ]]
}

is_wsl() {
    grep -qi microsoft /proc/version 2>/dev/null
}

is_macos() {
    [[ "${OSTYPE:-}" == "darwin"* ]]
}

# ==========================================
# UI Helpers
# ==========================================
print_banner() {
    clear
    printf "${COLOR_BRAND}====================================================${COLOR_RESET}\n"
    printf "${COLOR_BRAND}  FJ™ Cybertronic Systems // IDT Management Suite   ${COLOR_RESET}\n"
    printf "${COLOR_BRAND}====================================================${COLOR_RESET}\n\n"
}

print_sep() {
    printf "${COLOR_DIM}----------------------------------------------------${COLOR_RESET}\n"
}

press_enter() {
    echo -e "\n${COLOR_CYAN}Press Enter to return to the management menu...${COLOR_RESET}"
    read -r
}

info()  { printf "${COLOR_GREEN}[✔] %s${COLOR_RESET}\n" "$*"; }
warn()  { printf "${COLOR_YELLOW}[!] %s${COLOR_RESET}\n" "$*"; }
fail()  { printf "${COLOR_RED}[-] %s${COLOR_RESET}\n" "$*" >&2; }

# ==========================================
# Package Installation
# ==========================================
SUDO=""
if [ "$(id -u)" -ne 0 ] && command -v sudo &>/dev/null; then
    SUDO="sudo"
fi

_APT_UPDATED=""

install_package() {
    local pkg_name="$1"

    # Termux: hard branch, no sudo, no fallthrough
    if is_termux; then
        pkg install -y "$pkg_name" || return 1
        return 0
    fi

    if command -v apt-get &>/dev/null; then
        if [ -z "$_APT_UPDATED" ]; then
            $SUDO apt-get update && _APT_UPDATED=1
        fi
        $SUDO apt-get install -y "$pkg_name"
    elif command -v dnf &>/dev/null; then
        $SUDO dnf install -y "$pkg_name"
    elif command -v pacman &>/dev/null; then
        $SUDO pacman -S --noconfirm "$pkg_name"
    elif command -v zypper &>/dev/null; then
        $SUDO zypper install -y "$pkg_name"
    elif command -v apk &>/dev/null; then
        $SUDO apk add "$pkg_name"
    elif command -v brew &>/dev/null; then
        brew install "$pkg_name"
    else
        fail "No supported package manager found for: $pkg_name"
        return 1
    fi
}

# ==========================================
# Dependency Check
# ==========================================
check_dependencies() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Verifying system dependencies across environment...${COLOR_RESET}\n"

    if is_termux; then
        info "Environment Detected: Termux (Android)"
    elif is_wsl; then
        info "Environment Detected: WSL2 (Windows Subsystem for Linux)"
    elif is_macos; then
        info "Environment Detected: macOS"
    else
        info "Environment Detected: Standard Linux Server / Workstation"
    fi

    local failures=0

    # Go
    if ! command -v go &>/dev/null; then
        warn "Go runtime not found. Attempting automated installation..."
        if install_package golang; then
            info "Go installed successfully."
        else
            fail "Automatic Go installation failed. Install manually: https://golang.org/dl/"
            failures=$((failures + 1))
        fi
    else
        local go_ver
        go_ver="$(go version 2>/dev/null || echo 'unknown')"
        info "Go runtime detected: $go_ver"
    fi

    # Git
    if ! command -v git &>/dev/null; then
        warn "Git not found. Installing git..."
        if install_package git; then
            info "Git installed successfully."
        else
            warn "Could not install git automatically."
            failures=$((failures + 1))
        fi
    else
        info "Git version control detected."
    fi

    echo
    if [ "$failures" -gt 0 ]; then
        fail "$failures dependency issue(s) unresolved."
        press_enter
        return 1
    fi
    info "All system dependencies verified successfully!"
    press_enter
}

# ==========================================
# Build
# ==========================================
build_app() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Building IDT Binary via Go & Makefile...${COLOR_RESET}\n"

    if ! command -v go &>/dev/null; then
        fail "Go runtime not found. Run option 1 to install dependencies."
        press_enter
        return 1
    fi

    if [ -f "Makefile" ]; then
        if ! make build; then
            fail "Build failed via Makefile."
            press_enter
            return 1
        fi
    else
        mkdir -p bin
        if ! go build -o bin/idt main.go; then
            fail "Build failed via direct go build."
            press_enter
            return 1
        fi
    fi

    if [ -f "bin/idt" ]; then
        sha256sum bin/idt > bin/idt.sha256
        sha256sum -c bin/idt.sha256
        info "SHA-256 Checksum generated: bin/idt.sha256"
    fi

    info "Build successful! Binary located at: ./bin/idt"
    press_enter
}

# ==========================================
# Global Install
# ==========================================
install_system_wide() {
    print_banner

    local target_dir
    if is_termux; then
        target_dir="$PREFIX/bin"
        echo -e "${COLOR_BOLD}[*] Installing IDT to $target_dir/idt...${COLOR_RESET}\n"
    else
        target_dir="/usr/local/bin"
        echo -e "${COLOR_BOLD}[*] Installing IDT System-Wide ($target_dir/idt)...${COLOR_RESET}\n"
    fi

    if [ ! -f "bin/idt" ]; then
        warn "Binary not found. Building first..."
        if ! command -v go &>/dev/null; then
            fail "Go runtime not found. Run option 1 first."
            press_enter
            return 1
        fi
        mkdir -p bin
        if ! go build -o bin/idt main.go; then
            fail "Build failed."
            press_enter
            return 1
        fi
    fi

    echo -e "${COLOR_DIM}Target installation path: $target_dir/idt${COLOR_RESET}"

    if [ -w "$target_dir" ]; then
        cp bin/idt "$target_dir/idt" && chmod 755 "$target_dir/idt"
    elif is_termux; then
        # Termux: no sudo exists. Fail cleanly.
        fail "$target_dir is not writable on Termux. Aborting."
        press_enter
        return 1
    elif [ -n "$SUDO" ]; then
        warn "Root/Administrator permissions required to write to $target_dir."
        $SUDO cp bin/idt "$target_dir/idt" && $SUDO chmod 755 "$target_dir/idt"
    else
        fail "$target_dir is not writable and sudo is unavailable."
        press_enter
        return 1
    fi

    if command -v idt &>/dev/null; then
        info "Successfully installed 'idt' globally! Run 'idt' from anywhere."
    else
        warn "Binary installed to $target_dir/idt but not on PATH."
    fi
    press_enter
}

# ==========================================
# Tests
# ==========================================
run_tests() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Executing Go Test Suite & Race Detector...${COLOR_RESET}\n"

    if ! command -v go &>/dev/null; then
        fail "Go runtime not found. Run option 1 first."
        press_enter
        return 1
    fi

    if ! go test -v -race ./...; then
        fail "Test suite encountered failures or race conditions."
        press_enter
        return 1
    fi

    info "All tests passed successfully with zero race conditions!"
    press_enter
}

# ==========================================
# Git Update
# ==========================================
update_git() {
    print_banner
    echo -e "${COLOR_BOLD}[*] Synchronizing with Remote Git Repository...${COLOR_RESET}\n"

    if [ ! -d ".git" ]; then
        fail "Not a git repository."
        press_enter
        return 1
    fi

    if git pull origin main 2>/dev/null; then
        info "Updated from origin/main."
    elif git pull origin master 2>/dev/null; then
        info "Updated from origin/master."
    else
        fail "Git synchronization failed. Check network or local conflicts."
        press_enter
        return 1
    fi

    info "Repository successfully updated to the latest remote commit!"
    press_enter
}

# ==========================================
# Run IDT
# ==========================================
run_idt() {
    if [ -f "./bin/idt" ]; then
        ./bin/idt
    elif command -v idt &>/dev/null; then
        idt
    elif [ -f "main.go" ]; then
        go run main.go
    else
        fail "Binary not found. Please build first (Option 2)."
        press_enter
    fi
}

# ==========================================
# Main Menu
# ==========================================
while true; do
    print_banner
    echo -e "${COLOR_CYAN}  SELECT MANAGEMENT OPERATION:${COLOR_RESET}\n"
    print_sep
    echo -e "  ${COLOR_BOLD}1)${COLOR_RESET} Check Dependencies & Environment (Termux/WSL2/Linux/macOS)"
    echo -e "  ${COLOR_BOLD}2)${COLOR_RESET} Build IDT Application Binary"
    if is_termux; then
        echo -e "  ${COLOR_BOLD}3)${COLOR_RESET} Install Globally ($PREFIX/bin)"
    else
        echo -e "  ${COLOR_BOLD}3)${COLOR_RESET} Install Globally System-Wide (/usr/local/bin)"
    fi
    echo -e "  ${COLOR_BOLD}4)${COLOR_RESET} Run Test Suite & Race Analyzer"
    echo -e "  ${COLOR_BOLD}5)${COLOR_RESET} Update from Remote Git Repository"
    echo -e "  ${COLOR_BOLD}6)${COLOR_RESET} Run IDT Command Center"
    echo -e "  ${COLOR_BOLD}7)${COLOR_RESET} Exit Management Suite"
    print_sep
    echo -ne "\n  ${COLOR_BRAND}Enter choice [1-7]: ${COLOR_RESET}"
    read -r choice

    case "$choice" in
        1) check_dependencies ;;
        2) build_app ;;
        3) install_system_wide ;;
        4) run_tests ;;
        5) update_git ;;
        6) run_idt ;;
        7)
            print_banner
            echo -e "${COLOR_BRAND} Thank you for using FJ™ Cybertronic Systems IDT Suite.${COLOR_RESET}\n"
            exit 0
            ;;
        *)
            fail "Invalid choice. Please select between 1 and 7."
            sleep 1.5
            ;;
    esac
done
