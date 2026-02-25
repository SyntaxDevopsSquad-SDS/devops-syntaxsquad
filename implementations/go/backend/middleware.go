package main

import (
    "context"
    "net/http"
)

type contextKey string

const userContextKey contextKey = "user"

func WithUser(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := getSessionUser(r)
        ctx := context.WithValue(r.Context(), userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

func getUserFromContext(r *http.Request) string {
    user, ok := r.Context().Value(userContextKey).(string)
    if !ok {
        return ""
    }
    return user
}