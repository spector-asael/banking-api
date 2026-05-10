import { emitter } from "./event-emitter.js";

const API_BASE = "http://localhost:4000/api";

// Helper to get token from storage
const getAuthHeader = () => {
    const token = localStorage.getItem("auth_token");
    return token ? { "Authorization": `Bearer ${token}` } : {};
};

export const AuthService = {
    async login(email, password) {
        try {
            const response = await fetch(`${API_BASE}/tokens/authentication`, {
                method: "POST",
                body: JSON.stringify({ email, password }),
                headers: { "Content-Type": "application/json" }
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || "Login failed");
            }

            // Save the token!
            localStorage.setItem("auth_token", data.authentication_token.token);
            emitter.emit("auth:success", data.authentication_token);
        } catch (error) {
            emitter.emit("app:error", error.message);
        }
    },
    logout() {
        localStorage.removeItem("auth_token");
        emitter.emit("auth:logged-out");
    }
};

export const HistoryService = {
    async fetchHistory(accountNumber, page = 1, pageSize = 10, sort = "-date") {
        try {
            // 1. Path must match your working Go route exactly
            const url = `${API_BASE}/history/${accountNumber}?page=${page}&page_size=${pageSize}&sort=${sort}`;

            const response = await fetch(url, {
                headers: { ...getAuthHeader() }
            });

            if (response.status === 401) {
                AuthService.logout();
                throw new Error("Session expired.");
            }

            // 2. Get text first to see WHAT is breaking the parse
            const rawText = await response.text();
            
            // 3. CLEANING: This mimics the Go Decoder's "laziness"
            // It finds the first '{' and the last '}' and ignores everything else
            const cleanJson = rawText.substring(rawText.indexOf('{'), rawText.lastIndexOf('}') + 1);
            
            const data = JSON.parse(cleanJson);

            if (!response.ok) throw new Error(data.error || "Fetch failed");

            emitter.emit("history:loaded", { 
                ...data, 
                accountNumber, 
                currentPage: page, 
                pageSize, 
                sort 
            });
        } catch (error) {
            console.error("DEBUG: Raw response that failed:", error);
            emitter.emit("app:error", "Server sent invalid data. Check Go logs for fmt.Println calls.");
        }
    }
};