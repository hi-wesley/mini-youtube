# Mini YouTube

![Screenshot](https://i.imgur.com/N5vo39l.jpeg)

This is a full-stack web application that mimics some of the core functionalities of YouTube. It allows users to upload, watch, and comment on videos. The application also provides AI-powered video summarization.

## Features

*   **User Authentication:** Secure user registration and login using Firebase Authentication.
*   **Video Upload:** Upload video files to the platform.
*   **Video Streaming:** Stream videos seamlessly with a modern video player.
*   **Commenting System:** Engage in discussions with a real-time commenting feature.
*   **AI-Powered Summarization:** Get quick summaries of video content.
*   **Containerized Deployment:** Easily set up and run the application using Docker.

## Tech Stack

### Frontend

*   **React:** A JavaScript library for building user interfaces.
*   **TypeScript:** A typed superset of JavaScript that compiles to plain JavaScript.
*   **Vite:** A fast build tool and development server for modern web projects.
*   **Tailwind CSS:** A utility-first CSS framework for rapid UI development.
*   **Axios:** A promise-based HTTP client for the browser and Node.js.

### Backend

*   **Go:** A statically typed, compiled programming language designed at Google.
*   **Gin:** A web framework written in Go.
*   **Firebase Admin SDK:** For backend integration with Firebase services.
*   **GORM:** The fantastic ORM library for Go.

### Database

*   **PostgreSQL:** A powerful, open-source object-relational database system.

### Deployment

*   **Docker:** A platform for developing, shipping, and running applications in containers.
*   **Google Cloud Run:** A fully managed serverless platform for containerized applications.

## Project Structure

```
.
├── backend
│   ├── cmd
│   │   └── server
│   │       └── main.go
│   ├── internal
│   │   ├── ai
│   │   ├── config
│   │   ├── db
│   │   ├── firebase
│   │   ├── handlers
│   │   ├── middleware
│   │   └── models
│   ├── Dockerfile
│   └── go.mod
├── frontend
│   ├── src
│   │   ├── api
│   │   ├── components
│   │   └── main.tsx
│   ├── Dockerfile
│   └── package.json
├── deploy
│   ├── cloudrun-fe.yaml
│   └── cloudrun.yaml
├── docker-compose.yml
└── README.md
```

## API Endpoints

The backend API provides the following endpoints:

*   `POST /api/signup`: Register a new user.
*   `POST /api/login`: Log in an existing user.
*   `GET /api/videos`: Get a list of all videos.
*   `GET /api/videos/:id`: Get a single video by ID.
*   `POST /api/videos`: Upload a new video.
*   `GET /api/videos/:id/comments`: Get all comments for a video.
*   `POST /api/videos/:id/comments`: Add a new comment to a video.
*   `GET /api/videos/:id/summary`: Get a summary of a video.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.