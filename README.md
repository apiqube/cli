# ApiQube CLI

**ApiQube CLI** is a powerful, extensible command-line tool designed for building, executing, and monitoring tests for modern microservice applications—from simple HTTP APIs to complex multi-container systems. It is part of the [ApiQube](https://github.com/apiqube) ecosystem.

---

## 🚀 Features

- **Plugin Architecture**: Easily extend the CLI with custom plugins and actions.
- **Test Execution Engine**: Define use-cases, test flows, and assertions using YAML manifests.
- **Docker-Native**: Seamlessly spin up containers, stub databases, and create isolated environments for each test.
- **Plan Builder**: Build and apply test execution plans, similar to `kubectl apply`.
- **Load Testing Support**: Stress-test your services with real-world use cases.
- **Live Metric Collection**: Integrates with Prometheus and supports metrics agents for Go, JavaScript, and Python.
- **Future Wails UI**: Desktop testing studio with visual flow editing and live dashboards (coming soon).
- **CI-Ready**: Easily integrates with GitHub Actions and GitLab CI for automated testing.
- **Interactive CLI**: Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for a modern terminal experience.

---

## 📦 Installation

### Prebuilt Binaries (Recommended)

Prebuilt releases will be available soon via GitHub Releases and `go-semantic-release`. For now, build from source:

```bash
git clone https://github.com/apiqube/cli.git
cd cli
task build
cp ./bin/qube.exe ~/bin/qube  # or any directory in your PATH
```

---

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

---

## 📝 Example

Create a YAML manifest describing your test plan, then apply it:

```bash
qube apply -f my-test-plan.yaml
```

Search for existing manifests:

```bash
qube search --filter "service=api"
```

Rollback to a previous manifest version:

```bash
qube rollback --id my-service --version 2
```

---

## 🌍 Roadmap

- [x] CLI core
- [x] YAML-driven testing flows
- [ ] Visual test plan editor
- [ ] Plugin marketplace
- [ ] Built-in Prometheus integration
- [ ] Web dashboard (with Wails)
- [ ] GitHub/GitLab CI integration

---

## 🤝 Contributing

Contributions are welcome! Please open issues or pull requests on [GitHub](https://github.com/apiqube/cli).

---

## 📄 License

This project is licensed under the MIT License.
