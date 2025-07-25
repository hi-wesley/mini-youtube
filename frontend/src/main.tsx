import './index.css';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route } from 'react-router-dom';

import AuthProvider from './components/AuthProvider';
import Login from './components/Login';
import ProtectedRoute from './components/ProtectedRoute';
import RedirectIfAuth from './components/RedirectIfAuth';
import Layout from './components/Layout';
import ProfilePage from './components/ProfilePage';
import VideoList from './components/VideoList';
import VideoUpload from './components/VideoUpload';
import VideoPage from './components/VideoPage';

const qc = new QueryClient();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={qc}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route
              path="/login"
              element={
                <RedirectIfAuth>
                  <Login />
                </RedirectIfAuth>
              }
            />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <Layout>
                    <VideoList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/upload"
              element={
                <ProtectedRoute>
                  <Layout>
                    <VideoUpload />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/watch/:id"
              element={
                <ProtectedRoute>
                  <Layout>
                    <VideoPage />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <Layout>
                    <ProfilePage />
                  </Layout>
                </ProtectedRoute>
              }
            />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  </React.StrictMode>
);

