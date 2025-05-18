# 📝 Collaborative Notes App (Go + Fiber)

A real-time collaborative note-taking application built with **Go**, designed to help users create, edit, and share notes in a seamless and interactive environment.

## 🔧 Features

### Phase 1 (Current Scope)
- ✅ User signup & login with JWT
- ✅ Secure password hashing (bcrypt)
- ✅ CRUD operations for notes
- ✅ Protected routes with middleware
- ✅ PostgreSQL DB integration

### Later Phases (Planned)
- 🔄 Real-time note collaboration via WebSockets
- 🧠 Background auto-save jobs (Go routines + Redis)
- 📢 Pub/Sub for cross-instance updates
- 🧪 Versioning and edit history
- 👥 Online user presence

## 🛠️ Tech Stack

| Layer       | Tech                          |
|-------------|-------------------------------|
| Language    | Go                            |
| Web Server  | [Fiber](https://gofiber.io)   |
| Database    | PostgreSQL                    |
| Auth        | JWT + bcrypt                  |
| Cache/Queue | Redis (planned)               |
| Dev Tools   | Docker, Postman, Air (hot reload), godotenv |
