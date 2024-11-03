# G-FRESH

G-FRESH is a cutting-edge online platform designed to simplify grocery shopping. This project serves as the backend for an e-commerce platform, built using Go with the Gin framework, PostgreSQL as the database, and hosted on AWS. G-FRESH enables users to browse and purchase products, manage their profiles, and more. The API architecture is meticulously structured, with dedicated routes tailored for both users and admins, ensuring a smooth, intuitive experience for each group.

## G-FRESH Key Features

-  *Comprehensive Search and Filter:* Easily find products with advanced search options, including filters by category, availability, and sorting by price, popularity, or relevance, ensuring a tailored shopping experience.

- *Profile and Order Management:* Users can update their profiles, track orders in real-time, and view their complete order history for improved shopping insights.

- *Secure Authentication:* Enhanced user security through OTP verification during signup and robust authentication protocols to protect account access.

- *Role-Based Access Control:* Tailored functionalities for users and admins ensure optimized interactions and management capabilities for each role.

- *Referral and Wallet System:* Users earn rewards through referrals and can manage wallet balances directly on the platform, creating an incentive-driven experience.

- *Exciting Offers and Coupons:* Enjoy product discounts, seasonal promotions, and redeemable coupons, adding value and affordability to purchases.

- *Request for Out-of-Stock Products:* Users can send requests for out-of-stock products, helping ensure availability and keeping them engaged with the platform.

- *Location-Based Delivery Charges:* Delivery charges are calculated based on the user’s location, providing transparency and fair pricing for all customers.

- *Payment Integration:* Secure, seamless payment processing powered by Razorpay, making checkout fast and reliable.

- *Efficient Note Sharing:*  Cloud-based note upload and sharing, supported by Cloudinary, enables easy access to learning materials.

- *Wallet and Transaction History:* Transparent tracking of all transactions, including referral rewards, purchases, and wallet balances, for full user control over account activities.


## Installation

To set up the project locally, follow these steps:

1. *Clone the Repository:*

     bash
    git clone https://github.com/muhammedshamil123/G-Fresh.git
    cd G-Fresh
    
2. *Set Up the Environment Variables:*

    Create a .env file in the root directory and add the following variables:

    bash
    
    HOST=localhost
    USER=your_database_username
    PASSWORD=your_database_password
    DBNAME=your_database_name
    PORT=5432
    SSL=your_database_sslmode
    ZONE=your_database_timezone
    JWTSECRET=your_jwt_secret_key
    ClientID=your_google_auth_client_id
    ClientSecret=your_google_oauth_client_secret
    oauthStateString=your_oauth_state_string
    RazorpayID=your_razorpay_key_id
    RazorpaySecret=your_razorpay_key_secret
    Api=your_api
    Mail=your_smtp_app_password


3. *Install Dependencies:*

    bash
    go mod tidy
    

4. *Run the Application:*

    bash
    go run .
    

## API Documentation

Detailed API documentation is available [here](https://documenter.getpostman.com/view/38498526/2sAY4xB2kr).

---