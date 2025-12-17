.PHONY: setup dev clean help

help:
	@echo "Kaloupile Development Commands"
	@echo "=============================="
	@echo ""
	@echo "  make setup    - Run first-time setup (creates KinD cluster)"
	@echo "  make dev      - Start Tilt development environment"
	@echo "  make clean    - Delete the KinD cluster"
	@echo "  make help     - Show this help message"

setup:
	@go run ./cmd/setup

dev:
	@tilt up

clean:
	@echo "🗑️  Deleting KinD cluster..."
	@kind delete cluster --name kaloupile-dev
	@echo "✅ Cluster deleted"
