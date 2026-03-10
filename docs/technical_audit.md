# Technical Audit: Legacy "WhoKnows" Codebase

This document identifies the technical debt and security risks found in the original Python implementation. As we migrate this project to **Go**, this audit justifies our architectural choices and DevOps workflow.



## Summary of Findings
The legacy application is built on an **End-of-Life (EOL)** runtime (Python 2.7) and contains several critical security vulnerabilities. Our primary focus for the migration is to move from "reactive" coding to a "security-by-design" approach using Go's type safety and modern DevOps practices.


## Critical Issues (High Priority)

| Problem | Why it's dangerous | Our Fix in Go |
| :--- | :--- | :--- |
| **SQL Injection** | Hackers can "trick" the database into giving away all user data just by typing code into a login box. | We use **Prepared Statements** that treat user input as harmless text, never as executable code. |
| **Weak Passwords** | The app uses "MD5" to hide passwords. Modern computers can crack MD5 almost instantly. | We use **Bcrypt**, which adds a "salt" (random data) to every password, making it impossible to bulk-crack. |
| **Hardcoded Secrets** | Security keys are written directly in the code. Anyone who sees the GitHub repo can see them. | We move all secrets to **Environment Variables** (managed via a `.env` file), keeping them out of the code. |
| **Bad DB Management** | The app opens a new "door" to the database every time a user clicks. It eventually crashes the server. | We use a **Connection Pool**, which keeps a few "doors" open and reuses them efficiently. |




## Structural & Logic Issues (Medium Priority)

### 1. Missing "Digital Signature" on Forms (CSRF)
**The Problem:** The website doesn't verify if a request (like "change my password") actually came from the user clicking a button on *our* site.

**The "Fake Request" Risk:** A hacker could send you a link to a "funny cat video" site. While you're watching, that site sends a hidden command to our app: *"Delete my account."* Because you are already logged in, the legacy Python app just sees your "logged-in" cookie and says, *"Okay, account deleted!"*



**The Go Solution (CSRF Tokens):**
In the Go version, we will implement **CSRF Middleware**. 
- **Secret Generation:** Every time a user loads a form, the server generates a unique, one-time secret token.
- **Verification:** When the user submits the form, the Go server checks if that exact token is present.
- **Rejection:** Since a hacker's site can't "guess" this one-time token, their fake requests will be automatically rejected.

### 2. Outdated Language (Python 2)
The original code uses a version of Python that stopped getting security updates in 2020. It is like running an old operating system—it is full of holes that will never be patched.


## Code Quality & Style (Low Priority)

### 1. The "Everything" File
All the logic—database, security, and website routing—is crammed into one single file (`app.py`). 
- **The fix:** In our Go version, we split the code into organized folders (Packages) like `/internal/auth` and `/internal/database` so it’s easier to maintain and test.

### 2. Frontend Layout
We are keeping the original **HTML IDs and Tags** (as required by the exam), but we are adding modern **Tailwind CSS** classes to the tags to make the site look modern without breaking the legacy structure.


## Our DevOps Strategy
To move from this "Legacy Mess" to a modern Go app, we are using:
- **Docker:** To make sure the app runs the same on our laptops as it does on the server.
- **CI/CD:** To automatically test our code for security flaws (like the ones listed above) every time we push to GitHub.
- **12-Factor App Principles:** Ensuring our app is portable, scalable, and secure by default.
