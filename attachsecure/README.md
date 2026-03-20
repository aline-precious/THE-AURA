# AttachSecure

A next-generation **Attachment Style & Relationship Growth** platform built in Go, deployable to Vercel.

## Features

- **Attachment Assessment** — 7-question quiz that identifies your dominant attachment style
- **Dynamic Analyzer** — maps interaction patterns between two attachment styles (e.g. the Anxious-Avoidant trap)
- **AI Communication Coach** — translates messages based on your attachment style to reduce defensiveness
- **Progress Dashboard** — tracks your movement toward Earned Security with security score, mood trends, and check-in streaks
- **Daily Check-ins** — mood logging with trigger detection and style-specific grounding responses
- **Full PRD** — the complete Product Requirements Document is served at `/prd`

## Tech Stack

- **Language:** Go 1.21
- **Router:** gorilla/mux
- **Sessions:** gorilla/sessions (cookie-based)
- **Templates:** Go html/template (server-rendered)
- **Frontend:** Vanilla JS + CSS (no framework)
- **Deployment:** Vercel serverless via `api/index.go`

---

## Deploy to Vercel (from GitHub)

### 1. Push to GitHub

```bash
git init
git add .
git commit -m "initial commit"
git remote add origin https://github.com/YOUR_USERNAME/attachsecure.git
git push -u origin main
```

### 2. Import to Vercel

1. Go to [vercel.com](https://vercel.com) → **Add New Project**
2. Select your GitHub repository
3. Vercel auto-detects the `vercel.json` — no extra config needed
4. Add one environment variable:
   - `SESSION_SECRET` → any random 32+ character string (e.g. `openssl rand -base64 32`)
5. Click **Deploy**

That's it. Vercel builds and deploys automatically on every push to `main`.

---

## Run Locally

```bash
# Install dependencies
go mod tidy

# Run the dev server
go run cmd/server/main.go

# Open http://localhost:8080
```

---

## Project Structure

```
attachsecure/
├── api/
│   └── index.go          # Vercel serverless entry point (all logic inlined)
├── cmd/
│   └── server/
│       └── main.go       # Local dev server entry point
├── internal/
│   ├── ai/
│   │   └── coach.go      # Message translation, dynamic analysis, prompts
│   ├── handlers/
│   │   └── handlers.go   # HTTP route handlers
│   └── models/
│       └── models.go     # Data types, questions, style definitions
├── static/
│   ├── css/
│   │   └── main.css      # Full stylesheet
│   └── js/
│       └── main.js       # Client-side interactions
├── templates/
│   ├── layout.html       # Base layout with nav + footer
│   ├── home.html
│   ├── quiz.html
│   ├── result.html
│   ├── dashboard.html
│   ├── coach.html
│   ├── checkin.html
│   ├── about.html
│   └── prd.html
├── go.mod
├── go.sum
├── vercel.json           # Vercel deployment config
└── README.md
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `SESSION_SECRET` | Yes (production) | 32+ char secret for cookie signing |
| `PORT` | No | Port for local server (default: 8080) |

---

## Pages

| Route | Description |
|---|---|
| `/` | Home / landing page |
| `/quiz` | 7-question attachment assessment |
| `/result` | Your attachment style result |
| `/dashboard` | Progress dashboard |
| `/coach` | AI communication coach + dynamic analyzer |
| `/checkin` | Daily mood check-in |
| `/about` | About attachment theory |
| `/prd` | Full Product Requirements Document |

---

## Roadmap (v2)

- [ ] Persistent user accounts (PostgreSQL + magic-link auth)
- [ ] LLM-powered coach via Anthropic/OpenAI API
- [ ] Partner linking for couples
- [ ] Therapist dashboard with client management
- [ ] Push notifications for daily check-ins
- [ ] HIPAA compliance infrastructure

---

## License

MIT
