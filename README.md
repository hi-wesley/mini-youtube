# mini-youtube

![Screenshot](https://i.imgur.com/N5vo39l.jpeg)

This is a full-stack web application that aims to replicate some of the core functionalities of YouTube. It
  consists of a Go backend and a React frontend.

  Backend

  The backend is written in Go and uses the Gin web framework. Its main responsibilities are:

   * API: Providing a RESTful API for the frontend to interact with.
   * Authentication: Handling user authentication using Firebase.
   * Database: Using a PostgreSQL database with the GORM library to store information about videos, comments,
     and users.
   * Video Storage: Storing uploaded video files in Google Cloud Storage.
   * AI Summarization: Using Google's Vertex AI to generate summaries of the videos.
   * Real-time Comments: Leveraging WebSockets for a real-time commenting feature on video pages.

  Frontend

  The frontend is a single-page application built with React and TypeScript. Its key features are:

   * Video Uploading: A dedicated page for users to upload new videos.
   * Video Playback: A page to watch videos, which includes a video player and a comment section.
   * Routing: Uses React Router to handle navigation between the upload and watch pages.
   * Data Fetching: Uses TanStack Query (React Query) to manage data fetching from the backend, providing a
     smooth user experience.
   * Authentication: Integrates with the backend's authentication system to manage user sessions.

  Deployment

  The project is configured for deployment on Google Cloud Run, with separate services for the frontend and
  backend, as indicated by the deploy/cloudrun-fe.yaml and deploy/cloudrun.yaml files. The ci-cd.yaml file in
  the .github/workflows directory suggests a continuous integration and deployment pipeline is set up using
  GitHub Actions.
