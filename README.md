# 🛍️ E-Commerce Clothing Platform

A full-featured **e-commerce web application** tailored for selling **clothing**, built using the Go programming language. This project includes a clean admin panel, responsive user frontend, and secure backend services with JWT authentication.

---

## 🚀 Tech Stack

### 💻 Backend
- [Gin](https://github.com/gin-gonic/gin) – High-performance HTTP web framework
- [GORM](https://gorm.io/) – ORM for database interaction
- [JWT](https://jwt.io/) – Secure, stateless authentication
- [html/template](https://pkg.go.dev/html/template) – Server-side rendering with Go templates

### 🎨 Frontend
- HTML templates rendered from Go
- [Bootstrap 5](https://getbootstrap.com/) for UI components and responsiveness

---

## ✨ Key Features

### 👤 User Panel
- ✅ Register/Login with secure JWT-based sessions
- 🔐 OTP email verification and password reset
- 🛍️ Browse clothing products with filters (category, price, size)
- ❤️ Wishlist support
- 🛒 Cart and checkout functionality
- 📦 Order history and invoices

### 🛠️ Admin Panel
- 📊 Dashboard with sales reports and trending items
- 👕 Add/Edit/Delete products and variants
- 🔎 Search orders by ID, user, or date
- 🚫 Block/unblock users

---

## 📁 Project Structure

```text
.
├── DB/                 # Create DB instance
├── user/               # Route handlers User
├── admin/              # Route handlers Admin
├── helper/             # helper functions (JWT,OTP)
├── models/             # GORM models (User, Product, Order, etc.)
├── routes/             # API routing and middleware
├── templates/          # HTML templates rendered by Go
├── static/             # Static assets (CSS, JS, images)
├── utils/              # Helper utilities (html_helper,password)
├── go.mod              # Go module file
├── go.sum              # Go dependencies checksum
└── main.go             # Application bootstrap

```
---

## ⚙️ Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL (or any SQL DB)
- Git

### Installation

```bash
# Clone the repo
git clone [https://github.com/Varunjp/Golang-e-commerce.git]

# Set up environment variables
cp .env.example .env

# Install dependencies
go mod tidy

# Run the app
go run main.go
```

## 🔐 Environment Variables

Create a `.env` file in the root directory and add the following:

```env
PORT=8080                         # Port to run the server
dns                               # Database connection with username and password
ClientID                          # GoogleOauthID
ClientSecret                      # GoogleOauthSecret
GoogleRedirectURL                 # Redirect URL from Oauth
JWT_SECRET=your_jwt_secret        # Secret key for JWT generation
RAZORPAY_KEY_ID                   # Razorpay payment ID
RAZORPAY_KEY_SECRET               # Razorpay secret key
Email=noreply@project.com         # Sender email
Password=email_password           # Sender email password or app password

```

## 🙋‍♂️ Author

Created by Varun Jp
