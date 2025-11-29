import axios from "axios";

const api = axios.create({
  // PUT YOUR RENDER BACKEND URL HERE
  // example: "https://vantro-backend.onrender.com"
  baseURL: "https://YOUR-BACKEND-URL.onrender.com",
});

// Attach JWT token from localStorage
api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

export default api;
