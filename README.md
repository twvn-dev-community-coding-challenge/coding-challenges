# Coding Challenge 2026

Welcome to the TWVN Coding Challenge! Whether you're a seasoned developer or just starting your journey, this event is designed for **YOU**. This is a safe, fun space to experiment with new technologies, learn from your peers, and grow your skills — all while having a great time!

> **New to coding challenges?** Don't worry! We welcome all skill levels. Partial solutions, learning attempts, and "I tried something new" are all celebrated here.

---

## 📋 Table of Contents

1. [🌟 Why Join?](#why-join)
2. [🚀 Quick Start: How to Join](#quick-start-how-to-join)
3. [📏 Rules & Guidelines](#rules--guidelines)
4. [📤 How to Submit Your Code](#how-to-submit-your-code)
5. [🏆 Scoring, Leaderboard & Prizes](#scoring-leaderboard--prizes)
6. [🎤 Demo Day](#demo-day)
7. [💡 Tips for First-Timers](#tips-for-first-timers)
8. [❓ FAQ](#faq)
9. [💬 Communication & Contact](#communication--contact)

---

## 🌟 Why Join?

- 🚀 **Learn by doing** — Try that framework or pattern you've been curious about
- 🤝 **Connect with peers** — Work with people across teams you might not meet otherwise
- 🏆 **Build confidence** — No pressure, just growth and friendly competition
- 🎉 **Have fun** — Seriously, this is meant to be enjoyable!

### What You'll Get

- Knowledge Badges for your profile (e.g., 🟢 _TDD Specialist_, 🟢 _Clean Architecture_)
- A chance to present to the dev community on Demo Day
- **Team vouchers** for 1st, 2nd, and 3rd place at **each** coding challenge
- **Season Champion** — highest Season Score wins a **Mac mini M4** at the end of the season

---

## 🚀 Quick Start: How to Join

| Step                  | What to Do                                                                                                                                |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **1. Register**       | Drop a message in `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))                                  |
| **2. Form a Team**    | Join solo, form a team of 2-3 people, or ask in the channel for teammates                                                                 |
| **3. Attend Kickoff** | Get the challenge brief, repo access, and timeline                                                                                        |
| **4. Code & Learn**   | Spend 2-3 hours/week coding. Ask questions in the channel anytime                                                                         |
| **5. Submit**         | Fork the repo, implement in your fork, then open a Pull Request back to the official repo (see [How to Submit](#how-to-submit-your-code)) |
| **6. Present**        | Demo your solution on Demo Day and celebrate!                                                                                             |

---

## 📏 Rules & Guidelines

### 👥 Teams

- **Solo or teams of 1-3** — your choice
- **Pairing encouraged:** mix senior + junior intentionally
- Rotate pairs every challenge for variety

### ⏱️ Time Commitment

- **2-3 hours per week** over a 2 week coding period
- Partial solutions are welcome — "good enough" beats "perfect"

### ✅ What's Allowed

| ✅ Allowed                                                    | ❌ Not Allowed                  |
| ------------------------------------------------------------- | ------------------------------- |
| Any language/framework (unless specified)                     | Client data or proprietary code |
| AI tools (Copilot, ChatGPT, Cursor) — **bonus if documented** | Over-engineering beyond scope   |
| Open-source libraries                                         |                                 |

---

## 📤 How to Submit Your Code

You submit using a **fork-based workflow**:

1. The organizers create **one official challenge repository**.
2. Teams **fork** it and implement the solution in their own fork.
3. Teams submit by opening a **Pull Request back to the official repo** (or share the fork link as a fallback).

Follow the steps below in order.

### 🍴 Step 1: Fork the Official Repository

- Fork the official repo to your own GitHub account/org.
- All your work happens in your fork (not in the official repo).

### 💻 Step 2: Clone Your Fork

- Clone your fork locally.
- Add the upstream remote (optional but recommended) so you can pull updates.

### 🔀 Step 3: Create Your Working Branch (Recommended)

Avoid committing directly to your fork's `main`. Create a working branch right after you clone.

**Branch naming format (keep this):**

```
coding-challenge-{number}/<your-team-name>
```

**Examples:** `coding-challenge-1/rustaceans`, `coding-challenge-1/stevelam`

### 📁 Step 4: Add Your Code to Your Team Folder

Place **all** your code under your team's dedicated folder:

```
coding-challenge-{number}/<your-team-name>/
```

**Examples:**

- `coding-challenge-1/rustaceans/` — Team folder for the Rustaceans team
- `coding-challenge-1/stevelam/` — Individual folder for Steve Lam

**Your folder should contain:**

- All source code
- `README.md` (use the template below)
- Tests
- Any configuration files

#### README Template (for your team folder)

```markdown
# [Your Team Name] - [Challenge Name]

## Team Members

- Developer 1 (role/team)
- Developer 2 (role/team)

## How to Run

[Step-by-step instructions]

## Approach

[Brief explanation of your solution]

## Design Decisions

- Why did you choose this tech stack?
- What trade-offs did you make?
- What patterns did you use?

## Challenges Faced

[What was hard? How did you overcome it?]

## What We Learned

[New skills, technologies, or insights]

## With More Time, We Would...

[Nice-to-haves you didn't implement]

## AI Tools Used (if any)

[Which tools? How did they help?]
```

### 🔃 Step 5: Open a Pull Request to the Official Repo (Required) ⚠️

**Your PR is your official submission.** This is mandatory when possible.

1. Push your branch to **your fork**
2. Open a Pull Request from `your-fork:your-branch` into `official-repo:main`
3. Fill out the PR description using the template below
4. Request review from the challenge organizers or mentors
5. **Do not merge** — Organizers will review all PRs before Demo Day

**Fallback (only if PR is not possible):** Share the link to your fork (and the branch) with the organizers.

**PR Title Format:**

```
[Challenge #X] Team Name - Brief Solution Description
```

**PR Description Template:**

```markdown
## Team Information

- **Team Name**: [Your Team Name]
- **Members**:
  - [Name 1] - [Role/Team]
  - [Name 2] - [Role/Team]

## Solution Overview

[Brief description of your approach - 2-3 sentences]

## Key Design Decisions

- [Decision 1 and why]
- [Decision 2 and why]
- [Design patterns used]

## Technologies & Tools

- Language/Framework: [e.g., Python, Node.js, Rust]
- Key Libraries: [List main dependencies]
- AI Tools Used: [e.g., GitHub Copilot, ChatGPT - if any]

## How to Run

[Step-by-step instructions to run your code]

## How to Test

[Instructions to run tests if available]

## Trade-offs & Limitations

[What you optimized for and what was sacrificed]

## With More Time, We Would...

[Features or improvements you'd add]
```

---

## 🏆 Scoring, Leaderboard & Prizes

> **TL;DR** — Each challenge is scored out of 100 (+ bonus). Points are tracked per individual across each season of **four** coding challenges; your **Season Score** uses your **top three** challenge scores only. The highest Season Score wins the **Special Prize** (Mac mini M4). **Teams** placing 1st, 2nd, and 3rd at each challenge still receive **vouchers** for that round.

### How Points Work for Teams vs Individuals

| Model            | How It Works                                                                                               |
| ---------------- | ---------------------------------------------------------------------------------------------------------- |
| **Individual**   | You code alone. Your score goes directly to your profile.                                                  |
| **Team (max 3)** | The team is graded as one unit. Every member receives the full team score. Bonus points apply to everyone. |

All points are tracked **per individual** for the leaderboard, regardless of whether you join solo or in a team.

---

### Per-Challenge Scoring (Max 100 + Bonus)

#### A. Core Execution & Design — 70 Points

This evaluates the functional correctness and architectural quality of your submission. Each challenge defines its own detailed breakdown within this block (see the specific challenge document).

#### B. Tests & Reliability — 15 Points

| Criteria                                                    | Points |
| ----------------------------------------------------------- | ------ |
| **Coverage** — Tests covering the core logic                | 10     |
| **Resilience** — Tests for edge cases, errors, empty states | 5      |

#### C. Documentation & Communication — 15 Points

| Criteria                                                         | Points |
| ---------------------------------------------------------------- | ------ |
| **README** — Clear setup instructions and trade-off explanations | 5      |
| **Presentation** — Ability to explain _why_ decisions were made  | 10     |

---

### Bonus Points (Up to +25)

These let you exceed 100 and are the fast track to winning the Special Prize.

| Bonus                | Points    | How to Earn                                                                      |
| -------------------- | --------- | -------------------------------------------------------------------------------- |
| **TDD Approach**     | +5        | Commit history proves tests written _before_ implementation (Red-Green-Refactor) |
| **Junior-Led Demo**  | +5        | The most junior team member presents the solution                                |
| **Community Assist** | +5        | Provide constructive peer reviews on other participants' PRs                     |
| **Creative Bonus**   | up to +10 | Subjective — decided by the facilitator or community vote                        |

### Community Engagement Points (Demo Day)

Earned during Demo Day and added to your **Season Total**.

| Award                         | Points      | Details                                                       |
| ----------------------------- | ----------- | ------------------------------------------------------------- |
| **"Most Loved" Solution**     | +10         | Voted by the audience via live poll                           |
| **The Curious Mind**          | +2/question | Ask a meaningful technical question during Q&A (max +6/month) |
| **The Subject Matter Expert** | +2          | Give an exceptionally clear answer during your own Q&A        |

---

### The Long-Run Leaderboard (Seasons)

Your **Season Score** determines who wins the **Special Prize** for the season: a **Mac mini M4**.

Each **season** runs for **four** coding challenges. Your season score is the sum of your **three highest** challenge scores from that season—the fourth (lowest) score is dropped. If you complete **fewer than four** challenges, there is no extra score to drop: your season score is still the sum of your **three highest** scores among the challenges you entered (or all of your scores if you have fewer than three). There are no separate participation or placement bonuses on top of this.

**The "Best of 3 (of 4)" Rule:**

```
Season Score = (Top Score 1) + (Top Score 2) + (Top Score 3)
```

<details>
<summary><strong>Example: Alice vs Bob (click to expand)</strong></summary>

**Alice (The Consistent Performer)** — Completes all four challenges; one rough round is dropped.

| Challenge | Score  |
| --------- | ------ |
| C1        | 80     |
| C2        | **85** |
| C3        | **82** |
| C4        | 45     |

- Top 3 Scores: 85 + 82 + 80 = **247** (45 dropped)
- **Season Total: 247**

---

**Bob (Three Strong Rounds)** — Only takes part in three challenges this season—every score counts; there is no fourth score to drop.

| Challenge | Score  |
| --------- | ------ |
| C1        | **95** |
| C2        | **90** |
| C3        | **88** |
| C4        | —      |

- Top 3 Scores: 95 + 90 + 88 = **273**
- **Season Total: 273**

---

**Comparison**

|                  | Alice (4 challenges) | Bob (3 challenges) |
| ---------------- | -------------------- | ------------------ |
| **Top 3 Scores** | 247                  | 273                |
| **SEASON SCORE** | **247**              | **273**            |

> Takeaway: Having four scores lets you drop your weakest one. With only three scores, all of them count toward the same top-three cap—so one off week can’t be discarded.

</details>

---

### 🎁 Prizes & Recognition

**Each coding challenge:**

- 🎟️ **Team vouchers (1st, 2nd, 3rd)** — The top three **teams** for that challenge each receive vouchers according to placement (1st, 2nd, 3rd). This is unchanged: even with the season-long leaderboard, every challenge still has its own team podium and prizes.
- 🏅 Knowledge Badges for skills demonstrated (e.g., 🟢 _TDD Specialist_, 🟢 _Clean Architecture_)
- 🎤 Top 3 get the Demo Slot to present to the entire dev community

**End of each season:**

- 🏆 **The Season Champion (Special Prize)** — Highest **Season Score** wins a **Mac mini M4**

---

### 📅 Monthly Workflow

| Week         | Activity                                             |
| ------------ | ---------------------------------------------------- |
| **Week 1**   | Challenge Release & Registration                     |
| **Week 2-3** | Hacking & Development (PRs due by end of Week 3)     |
| **Week 4**   | Facilitator Review & Scoring                         |
| **Week 5**   | Demo Day & Winner Announcement (new challenge drops) |

---

## 🎤 Demo Day

Each team gets **10-15 minutes** to present:

| Time  | Activity                      |
| ----- | ----------------------------- |
| 1 min | Problem recap                 |
| 5 min | Live demo or code walkthrough |
| 3 min | Design decisions & trade-offs |
| 2 min | Lessons learned               |

You can also prepare:

- A 5-minute live demo OR recorded video
- Slides (optional, not required)

---

## 💡 Tips for First-Timers

### 🎯 Before You Start

1. **Read the challenge carefully** — Understand the requirements before coding
2. **Start simple** — Get something working, then improve it
3. **Don't aim for perfect** — "Good enough" is the goal
4. **Ask questions early** — Don't struggle alone for days

### 💻 During Development

1. **Time-box your work** — Set a timer for 2-3 hour blocks
2. **Commit frequently** — Small commits are easier to manage
3. **Document as you go** — Write down your decisions and trade-offs
4. **Test your solution** — Even basic tests count!

### 🎬 For Demo Day

1. **Practice your demo** — Run through it once before presenting
2. **Focus on learning** — Share what you learned, not just what you built
3. **Be honest** — "I tried X but it didn't work" is valuable!
4. **Have fun** — This isn't a job interview, it's a celebration

---

## ❓ FAQ

**Q: I'm a junior developer. Is this too advanced for me?**
A: Not at all! This event is designed for **all levels**. You'll be paired with others, and learning is valued as much as the solution itself. Many juniors have won past challenges!

**Q: What if I can't finish in time?**
A: Partial solutions are 100% welcome! Submit what you have and explain what you would do next.

**Q: What if my solution breaks during the demo?**
A: It happens! Explain what was supposed to happen. Judges value your understanding and design decisions over a perfect demo.

**Q: How much time should I really spend?**
A: Aim for **4-6 hours total** over 2 weeks. That's roughly 2 hours per week. Don't overwork — this should be fun!

**Q: Can I use code from previous projects?**
A: You can reuse your own patterns and approaches, but the solution should be built for this specific challenge. No copy-pasting entire projects.

**Q: What languages/frameworks are allowed?**
A: Unless the challenge specifies, use whatever you want! Want to try Rust? Go for it. Comfortable with Python? Perfect!

**Q: Do I need to use AI tools?**
A: Nope! It's optional. If you do use them (GitHub Copilot, ChatGPT, Cursor, etc.), document how they helped for bonus points.

**Q: What if I have a question not listed here?**
A: Ask in `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7)) or message the organizers. No question is too small!

---

## 💬 Communication & Contact

**📢 Main Channel:** `#Vietnam Dev Community` ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))
Use this for questions, teammate search, and general discussion.

**🆘 Need Help?**

- Senior devs volunteer as mentors — just ask!
- Stuck? Post in the channel, we're all here to help.

**📞 Organizers:**

- Steve Lam — <tai.lam@thoughtworks.com>
- Tom Tang — <tom.tang@thoughtworks.com>
- Technical Mentors — Check `#Vietnam Dev Community` pinned messages

**Have ideas to improve the event?** DM the organizers or share in the retro session.

---

## 🚀 Ready to Start?

1. ✅ Join the `#Vietnam Dev Community` space ([join here](https://chat.google.com/room/AAAAm0xrYkU?cls=7))
2. ✅ Mark the kickoff event in your calendar
3. ✅ Start thinking about who you'd like to team up with
4. ✅ Get excited — this is going to be fun! 🎉

**See you at the challenge!** 🚀
