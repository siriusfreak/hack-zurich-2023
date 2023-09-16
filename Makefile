.PHONY: run-backend
run-backend:
	cd backend && go run github.com/siriusfreak/hack-zurich-2023/backend/cmd

.PHONY: run-frontend
run-frontend:
	cd frontend && npm run start
