# Mini YouTube

[Link to Mini YouTube](https://wesleys-yt.web.app/)
![Screenshot](https://i.imgur.com/31W36lL.png)
![Screenshot](https://i.imgur.com/0ksz4jV.png)

A full-stack video sharing platform featuring video uploads, AI-powered summaries, commenting, and user authentication. Built with a Go backend and React frontend, designed for scalable deployment on Google Cloud.

---

## Technology Stack

| Area      | Technology                                                                                             |
| :-------- | :----------------------------------------------------------------------------------------------------- |
| **Backend** | [Go](https://golang.org/), [Gin](https://gin-gonic.com/), [GORM](https://gorm.io/), [PostgreSQL](https://www.postgresql.org/) |
| **Frontend**| [React](https://reactjs.org/), [TypeScript](https://www.typescriptlang.org/), [Vite](https://vitejs.dev/), [Tailwind CSS](https://tailwindcss.com/) |
| **Database**| [Supabase](https://supabase.io/) (PostgreSQL)                                                          |
| **Deployment**| [Docker](https://www.docker.com/), [Google Cloud Run](https://cloud.google.com/run), [Firebase Hosting](https://firebase.google.com/docs/hosting) |
| **Auth**    | [Firebase Authentication](https://firebase.google.com/docs/auth)                                       |
| **Storage** | [Google Cloud Storage](https://cloud.google.com/storage)                                               |
| **AI**      | [Google Vertex AI](https://cloud.google.com/vertex-ai) (Gemini) for video summarization              |

---

## Architecture

This project is a monorepo containing both backend and frontend applications. This separation allows for independent development, scaling, and deployment.

-   **Backend:** A Go API built with the Gin framework, deployed as a container to **Google Cloud Run**. It handles business logic, database interactions, and **orchestrates large file uploads** by generating secure, signed URLs.
-   **Frontend:** A React application built with Vite, deployed as a static site to **Firebase Hosting**. It provides the user interface and interacts with the backend API.
-   **Database:** A **Supabase PostgreSQL** instance stores user information, video metadata, comments, and likes.
-   **Authentication:** **Firebase Authentication** manages user identities and secures the backend API.
-   **File Storage:** Video files are uploaded **directly** from the client to **Google Cloud Storage** using secure signed URLs provided by the backend. This approach is highly scalable and avoids backend server limitations.
-   **AI Integration:** Videos are automatically processed with **Vertex AI (Gemini)** to generate summaries.

### Video Upload Flow

The architecture handles large file uploads following a modern, scalable pattern:

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Backend API
    participant GCS as Google Cloud Storage
    participant AI as Vertex AI
    participant Database

    User->>Frontend: Selects video file (up to 100MB)
    Frontend->>Backend API: 1. Request Signed URL (sends file metadata)
    Backend API-->>Frontend: 2. Returns one-time Signed URL
    Frontend->>GCS: 3. Uploads video file directly using Signed URL
    GCS-->>Frontend: Upload complete
    Frontend->>Backend API: 4. Finalize Upload (sends object name & title)
    Backend API->>Database: 5. Create video record
    Backend API->>AI: 6. Request video summary (async)
    Backend API-->>Frontend: Success
    AI-->>Backend API: 7. Returns summary (async)
    Backend API->>Database: 8. Update video with summary
```

---

## Database Schema

The application uses PostgreSQL with the following schema:

```mermaid
erDiagram
    users {
        string id PK "Firebase UID"
        string email UK "User email"
        string username UK "Case-insensitive unique"
        string avatar_url "Profile picture URL"
        timestamp created_at "Registration time"
    }
    
    videos {
        string id PK "UUID"
        string user_id FK "Owner's ID"
        string title "Video title (max 120)"
        text description "Video description"
        text thumbnail_url "Thumbnail image URL"
        string object_name "GCS object path"
        text summary "AI-generated summary"
        string summary_model "AI model used"
        bigint views "View count"
        timestamp created_at "Upload time"
    }
    
    comments {
        uint id PK "Auto-increment"
        string user_id FK "Commenter's ID"
        string video_id FK "Video ID"
        text message "Comment content"
        timestamp created_at "Comment time"
    }
    
    likes {
        string user_id PK,FK "User who liked"
        string video_id PK,FK "Liked video"
    }
    
    users ||--o{ videos : "uploads"
    users ||--o{ comments : "writes"
    users ||--o{ likes : "gives"
    videos ||--o{ comments : "has"
    videos ||--o{ likes : "receives"
```

### Schema Notes:
- **users**: Stores Firebase-authenticated users with unique, case-insensitive usernames
- **videos**: Video metadata with AI-generated summaries from Vertex AI
- **comments**: User comments on videos with real-time WebSocket support
- **likes**: Many-to-many relationship between users and videos (composite primary key)

---

## Features

-   **User Authentication:** Secure login and registration using Firebase Authentication with case-insensitive unique usernames
-   **Large Video Uploads:** Direct upload to Google Cloud Storage (up to 100MB) using signed URLs
-   **Video Playback:** Stream videos directly from Google Cloud Storage
-   **AI Summaries:** Automatic video summarization using Google's Gemini AI
-   **Commenting System:** Real-time comments on videos using WebSockets
-   **Like System:** Users can like and unlike videos
-   **Video Discovery:** Browse all uploaded videos with view counts

---

## API Endpoints

All endpoints are prefixed with `/v1`.

| Method | Endpoint                       | Description                                                              | Auth Required |
| :----- | :----------------------------- | :----------------------------------------------------------------------- | :------------ |
| `POST` | `/auth/check-username`         | Checks if a username is available (case-insensitive).                    | No            |
| `POST` | `/auth/register`               | Registers a new user with unique username.                               | Yes*          |
| `GET`  | `/profile`                     | Gets the profile of the current user.                                    | Yes           |
| `GET`  | `/videos`                      | Retrieves a list of all videos.                                          | No            |
| `POST` | `/videos/initiate-upload`      | Generates a secure signed URL for direct video upload to GCS.            | Yes           |
| `POST` | `/videos/finalize-upload`      | Confirms successful upload and creates the video record in the database. | Yes           |
| `GET`  | `/videos/:id`                  | Retrieves details for a single video.                                    | No            |
| `POST` | `/videos/:id/view`             | Increments the view count for a video.                                   | No            |
| `POST` | `/videos/:id/like`             | Toggles a like on a video.                                               | Yes           |
| `GET`  | `/videos/:id/comments`         | Retrieves all comments for a video.                                      | No            |
| `POST` | `/comments`                    | Creates a new comment on a video.                                        | Yes           |
| `GET`  | `/ws/comments?vid=<id>`        | Establishes a WebSocket connection for real-time comments.               | No            |

*Note: `/auth/register` requires a Firebase ID token in the Authorization header.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.