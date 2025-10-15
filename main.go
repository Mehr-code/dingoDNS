package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"os"
	"runtime"
	"time"
	"strings"
)

var providers = []string{
	"Shecan", "Radar", "Electro", "Begzar", "DNS Pro",
	"DynX", "403", "Google", "Cloudflare", "Clear all DNS",
}

var dnsServers = map[string][]string{
	"Shecan":          {"178.22.122.101", "185.51.200.1"},
	"Radar":           {"10.202.10.10", "10.202.10.11"},
	"Electro":         {"78.157.42.100", "78.157.42.101"},
	"Begzar":          {"185.55.226.26", "185.55.226.25"},
	"DNS Pro":         {"87.107.110.109", "87.107.110.110"},
	"DynX":            {"10.70.95.150", "10.70.95.162"},
	"403":             {"10.202.10.202", "10.202.10.102"},
	"Google":          {"8.8.8.8", "8.8.4.4"},
	"Cloudflare":      {"1.1.1.1", "1.0.0.1"},
}

func main() {
	fmt.Println("🌀 DINGO (mehr-code Edition)")
	fmt.Println("============================")

	showCurrentDNS()

	for {
		showMenu()
		fmt.Print("Select option (0 to exit): ")

		reader := bufio.NewReader(os.Stdin)
		input, _, _ := reader.ReadLine()
		choice := string(input)

		if choice == "0" {
			fmt.Println("👋 khodafez!")
			break
		}

		index := parseChoice(choice)
		if index < 1 || index > len(providers) {
			fmt.Println("❌ Invalid choice")
			continue
		}

		provider := providers[index-1]
		if provider == "Clear all DNS" {
	        clearAllDNS()
        } else {
		    updateDNS(provider)
		}
	}
}

func showCurrentDNS() {
	fmt.Println("\n📡 Current DNS:")
	if runtime.GOOS == "windows" {
		cmd := exec.Command("ipconfig", "/all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("❌ Could not read DNS automatically.")
			fmt.Println("👉 Please run `ipconfig /all` manually to see your DNS settings.")
			return
		}

		lines := strings.Split(string(output), "\n")
		fmt.Println("DNS Servers:")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DNS Servers") || strings.HasPrefix(line, "DNS Server") {
				fmt.Println(line)
			}
			// بعضی اوقات DNS دوم یا سوم تو خط بعدی میاد
			if strings.HasPrefix(line, " ") && len(line) > 0 && strings.Contains(line, ".") {
				fmt.Println(line)
			}
		}
	} else {
		file, err := os.ReadFile("/etc/resolv.conf")
		if err != nil {
			fmt.Println("Could not read resolv.conf:", err)
			return
		}
		fmt.Println(string(file))
	}
}

func showMenu() {
	fmt.Println("\nAvailable Providers:")
	for i, name := range providers {
		if servers, ok := dnsServers[name]; ok {
            fmt.Printf("  %2d) %-15s %v\n", i+1, name, servers)
        } else {
            fmt.Printf("  %2d) %-15s\n", i+1, name)
        }
	}
	fmt.Println("  0) Exit\n")
}

func parseChoice(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func clearAllDNS() {
	fmt.Println("🧹 Clearing all DNS entries (cross-platform)...")

	if runtime.GOOS == "windows" {
		ifaces, err := getActiveInterfaces()
		if err != nil {
			fmt.Println("❌ Could not detect interfaces:", err)
			return
		}
		if !isAdmin() {
			fmt.Println("❌ Administrator privileges required.")
			return
		}
		for _, iface := range ifaces {
			cmd := exec.Command("netsh", "interface", "ip", "delete", "dns", fmt.Sprintf("name=%q", iface), "all")
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("❌ Failed to clear DNS on %s: %v\nOutput: %s\n", iface, err, string(out))
				continue
			}
			fmt.Printf("✅ Cleared DNS on %s\n", iface)
		}
		fmt.Println("🎉 All DNS cleared on Windows!")

	} else if runtime.GOOS == "darwin" {
		ifacesCmd := exec.Command("networksetup", "-listallnetworkservices")
		out, err := ifacesCmd.Output()
		if err != nil {
			fmt.Println("❌ Could not list network services:", err)
			return
		}
		lines := strings.Split(string(out), "\n")
		for _, iface := range lines {
			if strings.TrimSpace(iface) == "" || strings.HasPrefix(iface, "An asterisk") {
				continue
			}
			cmd := exec.Command("networksetup", "-setdnsservers", iface, "Empty")
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("❌ Failed to clear DNS on %s: %v\nOutput: %s\n", iface, err, string(out))
				continue
			}
			fmt.Printf("✅ Cleared DNS on %s\n", iface)
		}
		fmt.Println("🎉 All DNS cleared on macOS!")

	} else {
		// assume Linux
		cmd := exec.Command("bash", "-c", "sudo sh -c '> /etc/resolv.conf'")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Failed to clear /etc/resolv.conf: %v\nOutput: %s\n", err, string(out))
			return
		}
		fmt.Println("🎉 /etc/resolv.conf cleared on Linux!")
	}
}



// isAdmin checks whether the current process runs with Administrator privileges
func isAdmin() bool {
	// Use PowerShell to check admin role; returns "True" or "False"
	cmd := exec.Command("powershell", "-Command",
		"([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")
	out, err := cmd.Output()
	if err != nil {
		// If PowerShell call fails, assume not admin (we will still try later and show error)
		return false
	}
	return strings.TrimSpace(strings.ToLower(string(out))) == "true"
}

// getActiveInterfaces tries to find active network interfaces (returns names).
// It first tries a PowerShell cmdlet (Get-NetAdapter), and if that fails falls back to parsing netsh output.
func getActiveInterfaces() ([]string, error) {
	// Preferred: PowerShell Get-NetAdapter (returns adapter names with Status "Up")
	psCmd := `Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Select-Object -ExpandProperty Name`
	cmd := exec.Command("powershell", "-Command", psCmd)
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var res []string
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln != "" {
				res = append(res, ln)
			}
		}
		if len(res) > 0 {
			return res, nil
		}
	}

	// Fallback: parse `netsh interface show interface`
	cmd = exec.Command("netsh", "interface", "show", "interface")
	out, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %v", err)
	}
	lines := strings.Split(string(out), "\n")
	var names []string
	// Skip header lines and parse afterwards. Each useful line ends with interface name.
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		// skip header if present
		if strings.HasPrefix(ln, "Admin") || strings.HasPrefix(ln, "----") || strings.HasPrefix(ln, "Interface") {
			continue
		}
		// Fields: Admin State  State    Type    Interface Name
		// We take the last column as the name.
		parts := strings.Fields(ln)
		if len(parts) >= 4 {
			// Interface name might be merged tokens — we reconstruct from 4th token onward
			name := strings.Join(parts[3:], " ")
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no active interfaces found")
	}
	return names, nil
}

// updateWindowsDNS applies the provided dnsServers to all active interfaces found.
// It checks admin privileges and prints clear error messages.
func updateWindowsDNS(provider string, dnsServers []string) {
	fmt.Printf("⚙️ Updating DNS to %s (Windows)...\n", provider)

	// Check admin
	if !isAdmin() {
		fmt.Println("❌ This operation requires Administrator privileges.")
		fmt.Println("👉 Please run the program as Administrator (right-click -> Run as administrator) and try again.")
		return
	}

	// Find active interfaces
	ifaces, err := getActiveInterfaces()
	if err != nil {
		fmt.Printf("❌ Could not detect active interfaces: %v\n", err)
		fmt.Println("👉 You can run `netsh interface show interface` to inspect interfaces manually.")
		return
	}

	fmt.Printf("Detected interfaces: %v\n", ifaces)

	// For each interface, set primary DNS and add additional DNSes
	for _, iface := range ifaces {
		nameArg := fmt.Sprintf("name=%q", iface)

		// 1️⃣ Delete all existing DNS entries
		delCmd := exec.Command("netsh", "interface", "ip", "delete", "dns", nameArg, "all")
		out, err := delCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Failed to delete existing DNS on %s: %v\nOutput: %s\n", iface, err, string(out))
			fmt.Println("👉 Make sure the interface name is correct and you have Administrator privileges.")
			return
		}
		fmt.Printf("🧹 Cleared existing DNS entries on %s\n", iface)

		// Primary (set)
		if len(dnsServers) == 0 {
			fmt.Println("❌ No DNS servers provided.")
			return
		}
	
		primary := dnsServers[0]
		setCmd := exec.Command("netsh", "interface", "ip", "set", "dns", nameArg, "static", primary)
		out, err = setCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Failed to set primary DNS %s on interface %s: %v\nOutput: %s\n", primary, iface, err, string(out))
			fmt.Println("👉 Make sure the interface name is correct and you have Administrator privileges.")
			return
		}
		fmt.Printf("✅ Set primary DNS %s on %s\n", primary, iface)

		// Remove any older secondary DNS entries? (Optional — here we just add additional entries)
		// Add second+ DNS entries
		for i := 1; i < len(dnsServers); i++ {
			dns := dnsServers[i]
			// index starts at 2 for secondary
			index := fmt.Sprintf("index=%d", i+1)
			addCmd := exec.Command("netsh", "interface", "ip", "add", "dns", nameArg, dns, index)
			out, err := addCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("❌ Failed to add DNS %s (index=%d) on %s: %v\nOutput: %s\n", dns, i+1, iface, err, string(out))
				fmt.Println("👉 You may need to remove existing DNS entries manually or check permissions.")
				return
			}
			fmt.Printf("✅ Added DNS %s (index=%d) on %s\n", dns, i+1, iface)
		}
	}

	fmt.Println("🎉 DNS update finished for detected interfaces.")
}


func updateLinuxDNS(provider string, dnsServers []string) {
	// ابتدا فایل /etc/resolv.conf را پاک می‌کنیم
	content := ""
	for _, dns := range dnsServers {
		content += fmt.Sprintf("nameserver %s\n", dns)
	}

	// نیاز به دسترسی sudo دارد
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | sudo tee /etc/resolv.conf > /dev/null", content))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to update /etc/resolv.conf: %v\nOutput: %s\n", err, string(out))
		return
	}

	fmt.Println("✅ Linux DNS Updated!")
}

func updateMacDNS(provider string, dnsServers []string) {
	// macOS: از networksetup استفاده می‌کنیم
	ifacesCmd := exec.Command("networksetup", "-listallnetworkservices")
	out, err := ifacesCmd.Output()
	if err != nil {
		fmt.Printf("❌ Could not list network services: %v\n", err)
		return
	}

	lines := strings.Split(string(out), "\n")
	for _, iface := range lines {
		if strings.TrimSpace(iface) == "" || strings.HasPrefix(iface, "An asterisk") {
			continue
		}

		// primary DNS
		cmdArgs := append([]string{"-setdnsservers", iface}, dnsServers...)
		cmd := exec.Command("networksetup", cmdArgs...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Failed to set DNS on %s: %v\nOutput: %s\n", iface, err, string(out))
			continue
		}
		fmt.Printf("✅ DNS set on %s\n", iface)
	}

	fmt.Println("🎉 macOS DNS Updated!")
}



func updateDNS(provider string) {
	fmt.Printf("⚙️ Updating DNS to %s...\n", provider)
	time.Sleep(3 * time.Second)

    if runtime.GOOS == "windows" {
		updateWindowsDNS(provider, dnsServers[provider])
	} else if runtime.GOOS == "darwin" {
		// macOS
		updateMacDNS(provider, dnsServers[provider])
	} else {
		// assume Linux
		updateLinuxDNS(provider, dnsServers[provider])
	}

	fmt.Println("✅ DNS Updated!")
}
