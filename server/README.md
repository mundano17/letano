# Authentication 

## Routes: 
- `/user/register` for registration of new users
- `/user/login` for login 
- `/user/delete` for deletion of users
- `/user/refresh` to rotate both refresh tokens and access token.

Logout is yet to be implemeneted

## Key Notes:
- The auth uses cookies for now (refer to the middleware and controller)
- In production, set secure to True in controller.
- access token expires every 5 mins
- refresh token expires ever 2 days.
- unit test for validPassword has been added, but more are yet to be added.