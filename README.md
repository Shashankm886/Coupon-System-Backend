Here is an overview of all the different components in this project, and how they interact with each other:

1. Client:
  - Represents users making requests to create or redeem coupons.
  - Routes: /coupons (GET, POST), /redeem (POST).

2. API Layer:
  - Gin Server: Receives requests and directs them to the appropriate controller methods.
  - Routes: Organized and defined in the routes package, forwarding requests to controllers.

3. Controller Layer:
  - CouponController: Handles business logic for creating and redeeming coupons.
  - Interfaces with the Service Layer to process the requests.

4. Routing Layer:
   - Concerned with routing management of the incoming requests to their respective controllers

5. Service Layer:
  - CouponService: Implements the core business logic.
  - Integrates the Grule Rule Engine for validating rules during coupon creation and redemption.
  - Connects with MongoDB collections to store or retrieve data.

6. Rule Layer:
   - Contains definition of Rules for Create and Redeem APIs
   - Separate rule layer provides clear demarcation of all the validations and checks belonging to business logic

7. MongoDB Database:
  - Collections:
    -> Coupons: Stores details about each coupon.
    -> Users: Contains user-specific data relevant to coupons.
    -> Orders: Holds order history, used for validation during redemption.

8. Environment Variables (.env):
  - Used to store MongoDB URI and other sensitive configurations securely.


In order to run the project:
- close the repo into your computer
- ensure you have all the necessary dependencies installed
- run the project in your localhost by using the following command in your terminal : go run main.go
