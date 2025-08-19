package middleware

import (
	"fmt"
	"net/http"

	pb "github.com/barathsurya2004/expenses/proto"

	"google.golang.org/grpc"
)

func CorsMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthorizationMiddleware(conn *grpc.ClientConn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		pClient := pb.NewUsersServiceClient(conn)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check the token in the database
			fmt.Println("Checking token in the database...") // Placeholder for actual token validation logic
			ctx := r.Context()
			res, err := pClient.CheckAuthToken(ctx, &pb.CheckAuthTokenRequest{
				AuthToken: token,
			})
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			fmt.Println("token:", token)
			if !res.GetIsValid() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
