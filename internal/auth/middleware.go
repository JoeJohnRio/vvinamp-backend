package auth

// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"strconv"

// 	"vvinamp/internal/pkg/jwt"
// 	userRepository "vvinamp/internal/repository/user"
// )

// var userCtxKey = &contextKey{"user"}

// type contextKey struct {
// 	name string
// }

// func Middleware() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			header := r.Header.Get("Authorization")

// 			// Allow unauthenticated users in
// 			if header == "" {
// 				next.ServeHTTP(w, r)
// 				return
// 			}

// 			//validate jwt token
// 			tokenStr := header
// 			username, err := jwt.ParseToken(tokenStr)
// 			if err != nil {
// 				http.Error(w, "Invalid token", http.StatusForbidden)
// 				return
// 			}

// 			// create user and check if user exists in db
// 			user := userRepository.User{Username: username}
// 			id, err := userRepository.GetUserIdByUsername(username)
// 			if err != nil {
// 				next.ServeHTTP(w, r)
// 				return
// 			}
// 			user.ID = strconv.Itoa(id)
// 			// put it in context
// 			ctx := context.WithValue(r.Context(), userCtxKey, &user)

// 			// and call the next with our new context
// 			r = r.WithContext(ctx)
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// // ForContext finds the user from the context. REQUIRES Middleware to have run.
// func ForContext(ctx context.Context) *userRepository.User {
// 	log.Println("joel123", userCtxKey)
// 	raw, _ := ctx.Value(userCtxKey).(*userRepository.User)
// 	return raw
// }
