# Product Module

## Overview

## API Endpoints
### Base Path
All products endpoints are relative to the domain base URL:
```
{BASE_URL}/products
```
---
### List Categories
**Endpoint:**
```
GET {BASE_URL}/products/categories
```
**Request Headers:**
```
Content-Type: application/json
```
**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": [
    {
      "id": "c1871497-e481-4ade-99a6-703d326fb44f",
      "name": {
        "en-US": "Experience"
      },
      "slug": {
        "en-US": "experience"
      },
      "description": {
        "en-US": "Experience"
      }
    },
    {
      "id": "34da4a50-99b6-4a74-b754-197e61a5da70",
      "name": {
        "en-US": "Tour"
      },
      "slug": {
        "en-US": "tour"
      },
      "description": {
        "en-US": "Tour"
      },
      "children": [
        {
          "id": "fc6a9433-9bbf-41e6-86f2-a44d9e7206be",
          "name": {
            "en-US": "Private"
          },
          "slug": {
            "en-US": "private"
          },
          "description": {
            "en-US": "Private"
          }
        },
        {
          "id": "5e5d0598-be35-4cd7-98ec-e0f117952918",
          "name": {
            "en-US": "Share"
          },
          "slug": {
            "en-US": "share"
          },
          "description": {
            "en-US": "Share"
          }
        },
        {
          "id": "aa3f5223-65f5-4ca9-8e51-6e41d6bfe0ad",
          "name": {
            "en-US": "Open"
          },
          "slug": {
            "en-US": "open"
          },
          "description": {
            "en-US": "Open"
          }
        }
      ]
    },
    {
      "id": "b5ac1434-cb5a-4f31-9bcd-5ee462778a66",
      "name": {
        "en-US": "Attraction"
      },
      "slug": {
        "en-US": "attraction"
      },
      "description": {
        "en-US": "Attraction"
      }
    }
  ],
  "trace_id": "a518ff60bbc7e1b52b779dd049a13790"
}
```

**Error Responses:**
---
### Retrieve Categories
**Endpoint:**
```
GET {BASE_URL}/products/categories/:uuid
```
**Request Headers:**
```
Content-Type: application/json
```
**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "id": "c1871497-e481-4ade-99a6-703d326fb44f",
    "name": {
      "en-US": "Experience"
    },
    "slug": {
      "en-US": "experience"
    },
    "description": {
      "en-US": "Experience"
    }
  },
  "trace_id": "901fe2defde1ea5a00c268b79ad8e49e"
}
```

**Error Responses:**
```json
{
  "success": false,
  "message": "category not found",
  "error_code": "CATEGORY_NOT_FOUND",
  "trace_id": "c3c82ca52687096f0506494a791ed381"
}
```
---
### Crate Category
**Endpoint:**
```
POST {BASE_URL}/products/categories
```
**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "parent_id": null,
  "name": {
    "en-US": "New Experience",
    "id-ID": "Pengalaman Baru" 
  },
  "slug": {
    "en-US": "new-experience",
    "id-ID": "pengalaman-baru"
  },
  "description": {
    "en-US": "New Experience",
    "id-ID": "Pengalaman Baru"
  }
}
```

**Request Schema:**

| Field                      | Type   | Required | Validation                         | Description                      |
|----------------------------|--------|----------|------------------------------------|----------------------------------|
| `parent_id`                | string | ❌ No     | uuid                               | UUID of parent category          |
| `name`                     | object | ✅ Yes    | object with key en-US and or id-ID | Localize name of category        |
| `slug`                     | object | ✅ Yes    | object with key en-US and or id-ID                              | Localize slug of category        |
| `description`              | object  | ❌ No    | object with key en-US and or id-ID                              | Localize description of category |

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "id": "019c4ebe-0c40-79b8-a349-d53e529d8524",
    "name": {
      "en-US": "New Experience",
      "id-ID": "Pengalaman Baru"
    },
    "slug": {
      "en-US": "new-experience",
      "id-ID": "pengalaman-baru"
    },
    "description": {
      "en-US": "New Experience",
      "id-ID": "Pengalaman Baru"
    }
  },
  "trace_id": "57e285cbcf1ae41e8e191960dc46cd28"
}
```

## Error Codes
### Entity Errors

| Code | Message          | HTTP Status | Description |
|------|------------------|-------------|-------------|
| `CATEGORY_NOT_FOUND` | category not found | 404 |  |
| `CATEGORY_ID_REQUIRED` | id is required                 | 409 |  |

## Database Schema
### Category Table

| Column         | Type      | Constraints | Description |
|----------------|-----------|-------------|-----------|
| `id`           | uuid      | PRIMARY KEY | Auto-generated booking ID |
| `parent_id`    | uuid      | NOT NULL | |
| `name`         | jsonb     | NOT NULL |  |
| `slug`         | jsonb     | NOT NULL|  |
| `description`  | jsonb     | NULL| |
| `icon_url`     | string    | NULL|  |
| `sort_order`   | string    | NULL | |
| `created_at`   | timestamp | NOT NULL | Unix timestamp (milliseconds) |
| `name_default` | string    | NULL | |
| `slug_default` | string    | NULL | |

## Business Rules