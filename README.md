# ApiQube CLI

**ApiQube** is a powerful, plugin-extensible CLI tool for building, executing, and monitoring tests for modern microservice applications — from simple HTTP APIs to complex multi-container systems.

Part of the [ApiQube](https://github.com/apiqube) ecosystem.

---

## 🚀 Features

-  **Plugin architecture** — extend Qube with custom actions
-  **Test execution engine** — define `use-cases`, test flows, and assertions via YAML
-  **Docker-native** — spin up containers, stub databases, and environments for each test
-  **Plan builder** — build and apply test execution plans (like `kubectl apply`)
-  **Load testing support** — stress your services with real use cases
-  **Live metric collection** — Prometheus integration, metrics agents for Go, JS, Python
-  **Future Wails UI** — desktop testing studio with visual flow and live dashboards
-  **CI-ready** — easily integrate with GitHub Actions / GitLab CI
-  **Interactive CLI** — powered by [Bubbletea](https://github.com/charmbracelet/bubbletea)

---

## 📦 Installation

### ✅ Prebuilt (recommended)

TBA via releases or `go-semantic-release`. For now:

```bash
git clone https://github.com/apiqube/cli.git
cd cli
task build
cp ./bin/qube.exe ~/bin/qube  # or any PATH directory
```

## 🧪 Usage
```bash
ApiQube is a powerful test manager for apps and APIs

Usage:
  qube [command]

Available Commands:
  apply       Apply resources from manifest file
  cleanup     Cleanup old manifest versions by its id
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  rollback    Rollback to previous manifest version
  search      Search for manifests using filters
  version     Print the version number

Flags:
  -h, --help   help for qube

Use "qube [command] --help" for more information about a command.
```

## 🌍 Roadmap
- [ ] CLI core
- [ ] YAML-driven testing flows
- [ ] Plugin marketplace
- [ ] Built-in Prometheus integration
- [ ] Web dashboard (with Wails)
- [ ] GitHub/GitLab CI integration
- [ ] Visual test plan editor
