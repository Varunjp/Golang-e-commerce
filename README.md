🛍️ E-Commerce Clothing Platform

A full-featured e-commerce web application built using Go (Golang) for the backend and HTML templates with Bootstrap for the frontend. This platform is tailored specifically for selling clothing online, offering features like user authentication, product browsing, cart management, order processing, and admin controls.

🔧 Tech Stack

🖥️ Backend

* Gin – High-performance HTTP web framework in Go

* GORM – ORM for interacting with the database

* JWT (JSON Web Tokens) – For stateless and secure user authentication

* Go Templates – Server-side rendering of HTML pages

🎨 Frontend

* HTML Templates – Dynamic rendering from the Go server

* Bootstrap 5 – For responsive and modern UI components

✨ Features

👤 User

* Register, login, and manage account

* Secure JWT-based authentication

* Browse and search clothing products

* Add/remove items from cart

* Place orders and view order history

* Forgot password & OTP verification support

🛒 Product

* Filter by category, price, and size

* View product variants (e.g., color/size options)

* Wishlist functionality

📦 Orders

* Checkout with multiple addresses

* Stock reduction on order placement

* Invoice generation

🛠️ Admin Panel

* Dashboard with monthly sales report

* Add/update/delete products and variants

* Manage users (view/block)

* View orders and search by user/date/order ID

🤝 Contributing

Contributions are welcome! Feel free to open issues or submit PRs for improvements, new features, or bug fixes.

📁 Project Structure

* /main.go         -> Entry point of the application

* /user            -> HTTP handlers for user routes

* /admin           -> HTTP handlers for admin routes

* /models          -> GORM models for DB schema

* /templates       -> HTML templates for rendering pages

* /static          -> Static assets (CSS, JS, images)

* /routes          -> Route definitions and middleware

* /utils           -> Helper functions (e.g., pagination, password)

* /middleware      -> All middleware configs

* /upload          -> Product images

* /helper          -> Helper configs (e.g., JWT,OTP)

