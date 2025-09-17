package tui

import (
	"fmt"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	runewidth "github.com/mattn/go-runewidth"
)

func (m model) updateLandingView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "h": // toggle header mode (full ASCII vs compact)
			m.forceCompactBanner = !m.forceCompactBanner
			return m, nil
		case "l": // Login
			m.state = viewAuth
			return m, nil

		case "r": // Register (redirect to auth for now)
			m.state = viewAuth
			return m, nil

		case "d": // Dev user login (temporary)
			m.state = viewAuth
			m.status = statusMessageStyle("Dev login: Use CLI command 'go run ./cmd/ossyne-cli dev users create -u testuser -e test@example.com' then 'auth login'")
			return m, nil

		case "b": // Browse projects (public)
			m.state = viewProjects
			m.loading = true
			m.status = statusMessageStyle("Loading public projects...")
			return m, m.apiClient.fetchUserProjectsCmd(1) // Load all projects

		case "c": // Create projects (protected)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to create projects.")
				m.state = viewAuth
				return m, nil
			}
			m.state = viewManageProjects
			m.loading = true
			m.status = statusMessageStyle("Loading your projects...")
			return m, m.apiClient.fetchUserProjectsCmd(m.loggedInUser.ID)

		case "m": // My contributions (protected)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to view your contributions.")
				m.state = viewAuth
				return m, nil
			}
			m.state = viewMyContributions
			return m, nil

		case "v": // Review contributions (protected)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to review contributions.")
				m.state = viewAuth
				return m, nil
			}
			m.state = viewReviewContributions
			return m, nil

		case "w": // My wallet (protected)
			if m.loggedInUser == nil {
				m.status = statusMessageStyle("Please login first to view your wallet.")
				m.state = viewAuth
				return m, nil
			}
			m.state = viewMyWallet
			return m, nil

		case "q": // Quit
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) updateAuthView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "l": // Login
			return m, func() tea.Msg { return startLoginFlowMsg{} }

		case "esc", "b": // Back to landing
			m.state = viewLanding
			return m, nil

		case "q": // Quit
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) viewLandingView() string {
	var b strings.Builder
	fullBannerLines := 6
	taglineLines := 2
	userCardLines := 0
	if m.loggedInUser != nil {
		userCardLines = 3
	}

	footerLines := 3
	available := m.height - 2
	neededWithFull := fullBannerLines + taglineLines + userCardLines + 15 + footerLines
	compactMode := false
	includeTagline := true
	if available > 0 && neededWithFull > available {
		includeTagline = false
		neededNoTagline := fullBannerLines + userCardLines + 15 + footerLines
		if neededNoTagline > available {
			compactMode = true
		}
	}
	if m.forceCompactBanner {
		compactMode = true
	}

	if m.height > 34 {
		b.WriteString("\n\n")
	} else if m.height > 28 {
		b.WriteString("\n")
	}
	useASCII := !compactMode
	if useASCII {
		ossyineHeader := []string{
			" ██████╗ ███████╗███████╗██╗   ██╗███╗   ██╗███████╗",
			"██╔═══██╗██╔════╝██╔════╝╚██╗ ██╔╝████╗  ██║██╔════╝",
			"██║   ██║███████╗███████╗ ╚████╔╝ ██╔██╗ ██║█████╗  ",
			"██║   ██║╚════██║╚════██║  ╚██╔╝  ██║╚██╗██║██╔══╝  ",
			"╚██████╔╝███████║███████║   ██║   ██║ ╚████║███████╗",
			" ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═══╝╚══════╝",
		}
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
		contentWidth := m.width - appStyle.GetHorizontalFrameSize()
		for _, line := range ossyineHeader {
			lw := runewidth.StringWidth(line)
			if contentWidth > lw {
				pad := (contentWidth - lw) / 2
				b.WriteString(strings.Repeat(" ", pad) + headerStyle.Render(line) + "\n")
			} else {
				b.WriteString(headerStyle.Render(line) + "\n")
			}
		}
	} else {
		// Compact header variant
		headerLines := []string{"OSSYNE", "Open Source Marketplace"}
		if m.width > 50 {
			headerLines[0] = "OSSYNE • Open Source Marketplace"
			headerLines = headerLines[:1]
		}
		headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
		contentWidth := m.width - appStyle.GetHorizontalFrameSize()
		for _, hl := range headerLines {
			lw := runewidth.StringWidth(hl)
			if contentWidth > lw {
				pad := (contentWidth - lw) / 2
				b.WriteString(strings.Repeat(" ", pad) + headerStyle.Render(hl) + "\n")
			} else {
				b.WriteString(headerStyle.Render(hl) + "\n")
			}
		}
	}
	b.WriteString("\n")

	if includeTagline {
		taglineText := "Connect maintainers with contributors • Earn bounties • Build reputation"
		taglineStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ea9b8ff")).
			Italic(true).
			Align(lipgloss.Center)
		if m.width > len(taglineText) {
			taglinePadding := (m.width - len(taglineText)) / 2
			taglinePaddingStr := strings.Repeat(" ", taglinePadding)
			b.WriteString(taglinePaddingStr + taglineStyle.Render(taglineText))
		} else {
			b.WriteString(taglineStyle.Render(taglineText))
		}
		if compactMode {
			b.WriteString("\n")
		} else {
			b.WriteString("\n\n")
		}
	}

	// User status card if logged in
	if m.loggedInUser != nil {
		userCardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Background(lipgloss.Color("#065F46")).
			Foreground(lipgloss.Color("#D1FAE5")).
			Padding(0, 1).
			Margin(0, 2, 0, 2)

		if m.width > 20 {
			userCardStyle = userCardStyle.Width(m.width - 8)
		}

		userCard := userCardStyle.Render(fmt.Sprintf("✓ Logged in as: %s", m.loggedInUser.Username))
		b.WriteString(userCard)
		b.WriteString("\n")
	}

	cardWidth := (m.width - 10) / 2
	if cardWidth < 30 {
		cardWidth = m.width - 8
	}

	authCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6"))
	if compactMode {
		authCardStyle = authCardStyle.Padding(0, 1).Margin(0, 1).Width(cardWidth)
	} else {
		authCardStyle = authCardStyle.Padding(1, 1).Margin(0, 1).Width(cardWidth)
	}

	authLines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E40AF")).Render("Authentication"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9ea9b8ff")).Render("Get started with secure access:"), "",
		"[l] Login with GitHub",
		"[r] Register Account",
		"[d] Dev User Info",
	}
	authContent := strings.Join(authLines, "\n")
	authCard := authCardStyle.Render(authContent)
	publicCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#059669"))
	if compactMode {
		publicCardStyle = publicCardStyle.Padding(0, 1).Margin(0, 1).Width(cardWidth)
	} else {
		publicCardStyle = publicCardStyle.Padding(1, 1).Margin(0, 1).Width(cardWidth)
	}

	publicLines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#047857")).Render("Public Access"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#9ea9b8ff")).Render("Explore without signing up:"), "",
		"[b] Browse Projects",
		"[p] View Public Tasks",
		"[i] Project Information",
	}
	publicContent := strings.Join(publicLines, "\n")
	publicCard := publicCardStyle.Render(publicContent)
	if m.loggedInUser != nil {
		protectedCardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED"))
		if compactMode {
			protectedCardStyle = protectedCardStyle.Padding(0, 1).Margin(0, 1).Width(cardWidth)
		} else {
			protectedCardStyle = protectedCardStyle.Padding(1, 1).Margin(0, 1).Width(cardWidth)
		}

		protLines := []string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6D28D9")).Render("Protected"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#9ea9b8ff")).Render("Member exclusive features:"), "",
			"[c] Create Projects",
			"[m] My Contributions",
			"[v] Review Contributions",
			"[w] My Wallet",
			"[s] Settings",
		}
		protectedContent := strings.Join(protLines, "\n")
		protectedCard := protectedCardStyle.Render(protectedContent)
		if cardWidth < m.width-20 {
			aLines := strings.Count(authCard, "\n")
			pLines := strings.Count(publicCard, "\n")
			if aLines < pLines {
				authCard += strings.Repeat("\n", pLines-aLines)
			} else if pLines < aLines {
				publicCard += strings.Repeat("\n", aLines-pLines)
			}
			row1 := lipgloss.JoinHorizontal(lipgloss.Top, authCard, publicCard)
			b.WriteString(row1 + "\n" + protectedCard)
		} else {
			b.WriteString(authCard + "\n" + publicCard + "\n" + protectedCard)
		}
	} else {
		infoCardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Background(lipgloss.Color("#FEF3C7")).
			Foreground(lipgloss.Color("#92400E"))
		if compactMode {
			infoCardStyle = infoCardStyle.Padding(0, 1).Margin(0, 1).Width(cardWidth)
		} else {
			infoCardStyle = infoCardStyle.Padding(1, 1).Margin(0, 1).Width(cardWidth)
		}

		infoContent := lipgloss.NewStyle().Bold(true).Render("💡 Members Only") + "\n\n" +
			"Login to unlock:\n" +
			"• Create & manage projects\n" +
			"• Contribute to tasks\n" +
			"• Earn bounty rewards\n" +
			"• Build your reputation"

		infoCard := infoCardStyle.Render(infoContent)

		if cardWidth < m.width-20 {
			aLines := strings.Count(authCard, "\n")
			pLines := strings.Count(publicCard, "\n")
			if aLines < pLines {
				authCard += strings.Repeat("\n", pLines-aLines)
			} else if pLines < aLines {
				publicCard += strings.Repeat("\n", aLines-pLines)
			}
			row1 := lipgloss.JoinHorizontal(lipgloss.Top, authCard, publicCard)
			b.WriteString(row1 + "\n" + infoCard)
		} else {
			b.WriteString(authCard + "\n" + publicCard + "\n" + infoCard)
		}
	}

	b.WriteString("\n")
	footerText := "[q] Quit • Press the highlighted keys to navigate"
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ea9b8ff")).
		Padding(1, 2).
		Align(lipgloss.Center)

	if m.width > len(footerText) {
		footerPadding := (m.width - len(footerText)) / 2
		footerPaddingStr := strings.Repeat(" ", footerPadding)
		b.WriteString(footerPaddingStr + footerStyle.Render(footerText))
	} else {
		b.WriteString(footerStyle.Render(footerText))
	}

	if m.status != "" && !strings.Contains(strings.ToLower(m.status), "welcome back") && !strings.Contains(strings.ToLower(m.status), "welcome to ossyne") {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#059669")).
			Background(lipgloss.Color("#D1FAE5")).
			Padding(0, 2).
			Margin(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981"))

		b.WriteString("\n\n")
		b.WriteString(statusStyle.Render(m.status))
	}

	return b.String()
}

func (m model) viewAuthView() string {
	var b strings.Builder

	title := titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render("🔐 Authentication Required")
	b.WriteString(title)
	b.WriteString("\n\n")

	instructions := "Please authenticate to access protected features."
	b.WriteString(instructions)
	b.WriteString("\n\n")

	options := []string{
		"[l] Login with GitHub - Opens browser for OAuth",
		"[esc] Back to Landing",
		"[q] Quit",
	}
	b.WriteString(strings.Join(options, "\n"))

	b.WriteString("\n\n")
	b.WriteString(formHelpStyle.Render(m.status))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}
func (m model) updateCreateProjectView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "p":
			if m.loggedInUser != nil {
				m.state = viewProjects
				m.loading = true
				m.status = statusMessageStyle("Loading your projects...")
				return m, m.apiClient.fetchUserProjectsCmd(m.loggedInUser.ID)
			}
			return m, nil
		case "esc", "b":
			m.state = viewLanding
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) viewCreateProjectView() string {
	var b strings.Builder
	title := titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render("🏗️ Create & Manage Projects")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loggedInUser != nil {
		b.WriteString(fmt.Sprintf("Welcome, %s! Here you can create and manage your projects.\n\n", m.loggedInUser.Username))

		options := []string{
			"Available Actions:",
			"• [p] View Your Projects",
			"• [n] Create New Project (Feature in development)",
			"• [t] Manage Project Tasks",
			"• [f] Fund Project Bounties",
			"",
			"Note: Full project creation UI coming soon!",
			"For now, use CLI: 'go run ./cmd/ossyne-cli project create'",
		}
		b.WriteString(strings.Join(options, "\n"))
	} else {
		b.WriteString("Authentication required to access this section.")
	}

	b.WriteString("\n\n[p] Your Projects • [esc] Back to Landing • [q] Quit")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m model) updateMyContributionsView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "b":
			m.state = viewLanding
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) viewMyContributionsView() string {
	var b strings.Builder
	title := titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render("📝 My Contributions")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loggedInUser != nil {
		b.WriteString(fmt.Sprintf("Welcome, %s! Track your contribution history here.\n\n", m.loggedInUser.Username))

		info := []string{
			"This section will show:",
			"• Your submitted pull requests",
			"• Contribution status (pending/accepted/rejected)",
			"• Earned bounties and rewards",
			"• Contribution ratings and feedback",
			"",
			"🚧 Full implementation coming soon!",
			"For now, use CLI: 'go run ./cmd/ossyne-cli task list-contributions'",
		}
		b.WriteString(strings.Join(info, "\n"))
	} else {
		b.WriteString("Authentication required to view contributions.")
	}

	b.WriteString("\n\n[esc] Back to Landing • [q] Quit")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m model) updateReviewContributionsView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "b":
			m.state = viewLanding
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) viewReviewContributionsView() string {
	var b strings.Builder
	title := titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render("🔍 Review Contributions")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loggedInUser != nil {
		b.WriteString(fmt.Sprintf("Welcome, %s! Review contributions to your projects here.\n\n", m.loggedInUser.Username))

		info := []string{
			"Review Panel Features:",
			"• View pending contributions to your projects",
			"• Accept or reject pull requests",
			"• Rate contributor performance",
			"• Manage bounty payouts",
			"• Leave feedback and suggestions",
			"",
			"🚧 Full review UI coming soon!",
			"For now, use CLI: 'go run ./cmd/ossyne-cli mentor accept-contribution'",
		}
		b.WriteString(strings.Join(info, "\n"))
	} else {
		b.WriteString("Authentication required to review contributions.")
	}

	b.WriteString("\n\n[esc] Back to Landing • [q] Quit")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m model) updateMyWalletView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "b":
			m.state = viewLanding
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) viewMyWalletView() string {
	var b strings.Builder
	title := titleStyle.Width(m.width - appStyle.GetHorizontalFrameSize()).Align(lipgloss.Center).Render("💰 My Wallet")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loggedInUser != nil {
		b.WriteString(fmt.Sprintf("Welcome, %s! Manage your earnings and payments here.\n\n", m.loggedInUser.Username))

		info := []string{
			"Wallet Features:",
			"• View total earnings from bounties",
			"• Track payment history",
			"• Manage withdrawal methods",
			"• View pending and completed payments",
			"• Transaction history and receipts",
			"",
			"Full wallet integration coming soon!",
			"For now, use CLI: 'go run ./cmd/ossyne-cli payment history'",
		}
		b.WriteString(strings.Join(info, "\n"))
	} else {
		b.WriteString("Authentication required to access wallet.")
	}

	b.WriteString("\n\n[esc] Back to Landing • [q] Quit")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}