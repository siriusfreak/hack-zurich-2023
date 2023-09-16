.PHONY: run-backend
run-backend:
	cd backend && go run github.com/siriusfreak/hack-zurich-2023/backend/cmd

.PHONY: run-frontend
run-frontend:
	cd frontend && npm run start


.PHONY: run-frontend-production
run-frontend-production:
	cd frontend && REACT_APP_AUTH_TOKEN="Basic " REACT_APP_STAGE=production npm run start
