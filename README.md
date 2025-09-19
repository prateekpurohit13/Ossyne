# OSSYNE: The CLI-First Open-Source DevEx

<div align="center">

![OSSYNE CLI Help Screen](ossyne-cli-help.png)

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![GitHub OAuth](https://img.shields.io/badge/Auth-GitHub_OAuth-181717?style=for-the-badge&logo=github)](https://docs.github.com/en/developers/apps/building-oauth-apps)
[![MySQL](https://img.shields.io/badge/Database-MySQL-4479A1?style=for-the-badge&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)

**🚀 A developer-centric platform for open-source collaboration**

[Installation](#️-installation--setup) • [Features](#-features--highlights) • [Getting Started](#-getting-started-with-the-tui) • [Contributing](#-contributing--future-scope)

</div>

---

## 🚀 Introduction

**OSSYNE** is an innovative, command-line interface (CLI) first DevEx platform designed to seamlessly connect open-source project maintainers with talented contributors. Inspired by the principles of platforms like Upwork but built from the ground up for developer workflows, OSSYNE empowers you to find, claim, contribute to, and manage open-source tasks directly from your terminal.

Whether you're a maintainer looking for skilled help or a developer seeking to build your portfolio and earn bounties, OSSYNE offers a developer-centric experience with:

* **🖥️ Native CLI & Interactive TUI:** Perform all actions from your terminal  
* **🔗 Git-Native Workflows:** Integration with Git concepts (issues ↔ PRs ↔ credits)  
* **✅ Automated Contribution Verification:** Streamlined verification of contributions via Git metadata  
* **🏆 Reputation & Mentorship Systems:** Build trust and foster growth within the community  
* **💰 Optional Bounties & Escrow:** Financial incentives for impactful contributions  

---

## ✨ Features & Highlights

### 💻 CLI-First & Interactive TUI

OSSYNE offers a powerful dual interface: a robust command-line interface for scripting and direct actions, and an intuitive Terminal User Interface (TUI) for an interactive, dashboard-like experience.

* **⚡ CLI for Power Users:** Execute specific commands for user management, project creation, task claims, and more  
* **🎛️ Interactive TUI:** Navigate your marketplace with a sleek, multi-panel interface. Browse projects and tasks, manage your contributions, and initiate workflows with simple keybindings  

<div align="center">

![OSSYNE Landing Page - Logged In](ossyne-landing-logged-in.png)
*OSSYNE TUI Dashboard - Clean, intuitive interface*

</div>

#### 🔐 1. Authentication
Log in securely using your GitHub account directly from the CLI or TUI.

#### 🌐 2. Browse & Contribute
Explore public projects and their tasks. If logged in, claim tasks, submit Pull Request links, and track your contributions.

<div align="center">

![OSSYNE Browse Public Projects](ossyne-browse-public-projects.png)  
*Browse and discover open-source projects*

![OSSYNE Project Tasks View](ossyne-project-tasks-view.png)  
*View available tasks with detailed information*

</div>

#### 🛠️ 3. Maintainer Tools
For project maintainers, OSSYNE provides dedicated tools to manage your open-source initiatives.

* **📋 Your Projects View:** See all projects you own  
  
  <div align="center">
  
  ![OSSYNE Your Projects View](ossyne-maintain-projects-dash.png)
  
  </div>

* **➕ Create New Project:** Launch a new open-source project  
  
  <div align="center">
  
  ![OSSYNE Create New Project Form](ossyne-create-new-project-form.png)
  
  </div>

* **📝 Create New Task:** Add new tasks with guided forms  
  
  <div align="center">
  
  ![OSSYNE Create New Task Form](ossyne-create-new-task-form.png)
  
  </div>

* **💎 Fund Bounties:** Attach incentives via a simulated escrow system  

---

### 🔐 Secure & Accountable

* **🔑 GitHub OAuth2:** Industry-standard authentication  
* **👥 Role-Based Access:** Maintainers control task creation and approvals  
* **💳 Simulated Payments:** Bounties processed with escrow logic  
* **⭐ Ratings System:** Contributions verified via Git metadata  

---

## ⚙️ Installation & Setup

### 📋 Prerequisites

* [**Go**](https://golang.org/dl/) (1.18 or newer)  
* [**Docker**](https://www.docker.com/get-started) and **Docker Compose**  
* [**`golang-migrate` CLI tool**](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)  

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 🔧 Setup Steps

#### 1. 📂 Clone the Repository

```bash
git clone https://github.com/your-username/ossyne.git
cd ossyne
```

#### 2. 🔐 Set up GitHub OAuth App

1. Go to your [**GitHub Developer Settings**](https://github.com/settings/developers)
2. Create a **New OAuth App** with:
   * **Application name:** `OSSYNE CLI Dev`
   * **Homepage URL:** `http://localhost:8080`
   * **Authorization callback URL:** `http://localhost:8080/auth/github/callback`
3. Copy your generated **Client ID** and **Client Secret**

#### 3. 🔧 Configure Environment Variables

Create a `.env` file in the root of your `ossyne` directory:

```env
# MySQL Database Configuration
DB_USER=ossyne_user
DB_PASSWORD=secretpassword
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=ossyne_db

# Server Configuration
SERVER_PORT=8080

# GitHub OAuth Configuration
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback
```

> **⚠️ Important:** Replace placeholders with your actual values.

#### 4. 🗄️ Start Database & Apply Migrations

```bash
# Start MySQL with Docker Compose
docker-compose up -d

# Apply database migrations
migrate -path migrations -database 'mysql://ossyne_user:secretpassword@tcp(127.0.0.1:3306)/ossyne_db?charset=utf8mb4&parseTime=True&loc=Local' up
```

#### 5. 🖥️ Run the OSSYNE Server

```bash
go run ./cmd/ossyne-server/main.go
```

#### 6. 🚀 Run the OSSYNE CLI / TUI

Open a **new terminal**:

```bash
# Run CLI commands
go run ./cmd/ossyne-cli/main.go help

# Launch the TUI
go run ./cmd/ossyne-cli/main.go
```

---

## 🚀 Getting Started with the TUI

1. **Launch:** Run `go run ./cmd/ossyne-cli/main.go`
2. **Welcome:** The Welcome Screen will appear
3. **Login:** Press `l` to log in with GitHub OAuth. Your browser opens → authorize app
4. **Relaunch:** `go run ./cmd/ossyne-cli/main.go` → "Welcome back, [your_username]!"
5. **Navigate:** Use **Arrow Keys, Enter, Esc**, and shortcuts (`b`, `p`, `n`, `f`, etc.)

> 💡 **Pro Tip:** Use keyboard shortcuts for faster navigation through the TUI!

---

## 🤝 Contributing & Future Scope

<div align="center">

**🌟 OSSYNE is evolving. Contributions welcome! 🌟**

[![Contributions Welcome](https://img.shields.io/badge/Contributions-Welcome-brightgreen.svg?style=for-the-badge)](CONTRIBUTING.md)

</div>

### 🛣️ Roadmap & Planned Features

- [ ] **👤 Full user registration UI**
- [ ] **🔧 Advanced maintainer tools** (review/reject contributions, team management)
- [ ] **📊 "My Contributions" & "My Wallet" views in the TUI**
- [ ] **💳 Real payment integration** (Stripe/PayPal)
- [ ] **🔍 Search & recommendation engine**
- [ ] **🔔 Notifications & alerts**
- [ ] **🎨 ASCII art splash screen**

### 🛠️ How to Contribute

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.

---

<div align="center">

**Made with ❤️ for the open-source community**

[![GitHub stars](https://img.shields.io/github/stars/prateekpurohit13/ossyne?style=social)](https://github.com/prateekpurohit13/ossyne/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/prateekpurohit13/ossyne?style=social)](https://github.com/prateekpurohit13/ossyne/network/members)

[⬆ Back to top](#ossyne-the-cli-first-open-source-marketplace)

</div>
