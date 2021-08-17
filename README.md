# GoScrapeRedmine

#### API
- POST: /signup
  - Body example:
   ```json
    {
        "name": "Pham Van A", // free string
        "email": "phamvana@gmail.com", // free string, unique in database
        "password": "password", // free string
        "role": "admin" // free string
    }
    ```
